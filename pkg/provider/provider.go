// Copyright 2016-2021, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
)

type xyzProvider struct {
	host     *provider.HostClient
	name     string
	version  string
	pkgSpec  *schema.PackageSpec
	metadata *APIMetadata
}

func makeProvider(host *provider.HostClient, name, version string, schemaBytes []byte,
	apiResourcesBytes []byte) (rpc.ResourceProviderServer, error) {
	uncompressed, err := gzip.NewReader(bytes.NewReader(schemaBytes))
	if err != nil {
		return nil, errors.Wrap(err, "expand compressed schema")
	}

	var pkgSpec schema.PackageSpec
	if err = json.NewDecoder(uncompressed).Decode(&pkgSpec); err != nil {
		return nil, fmt.Errorf("deserializing schema: %w", err)
	}

	var metadata APIMetadata
	uncompressed, err = gzip.NewReader(bytes.NewReader(apiResourcesBytes))
	if err != nil {
		return nil, errors.Wrap(err, "expand compressed metadata")
	}
	if err = json.NewDecoder(uncompressed).Decode(&metadata); err != nil {
		return nil, errors.Wrap(err, "unmarshalling resource map")
	}
	if err = uncompressed.Close(); err != nil {
		return nil, errors.Wrap(err, "closing uncompress stream for metadata")
	}

	// Return the new provider
	return &xyzProvider{
		host:     host,
		name:     name,
		version:  version,
		pkgSpec:  &pkgSpec,
		metadata: &metadata,
	}, nil
}

// CheckConfig validates the configuration for this provider.
func (p *xyzProvider) CheckConfig(_ context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	return &rpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig diffs the configuration for this provider.
func (p *xyzProvider) DiffConfig(_ context.Context, _ *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	return &rpc.DiffResponse{}, nil
}

// Configure configures the resource provider with "globals" that control its behavior.
func (p *xyzProvider) Configure(_ context.Context, _ *rpc.ConfigureRequest) (*rpc.ConfigureResponse, error) {
	return &rpc.ConfigureResponse{}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (p *xyzProvider) Invoke(_ context.Context, _ *rpc.InvokeRequest) (*rpc.InvokeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "StreamInvoke is not yet implemented")
}

// StreamInvoke dynamically executes a built-in function in the provider. The result is streamed
// back as a series of messages.
func (p *xyzProvider) StreamInvoke(_ *rpc.InvokeRequest, _ rpc.ResourceProvider_StreamInvokeServer) error {
	return status.Error(codes.Unimplemented, "StreamInvoke is not yet implemented")
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (p *xyzProvider) Check(_ context.Context, req *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	typ := resource.URN(req.GetUrn()).Type()

	if _, ok := p.pkgSpec.Resources[typ.String()]; !ok {
		return nil, fmt.Errorf("unknown resource type %q", typ)
	}
	return &rpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (p *xyzProvider) Diff(_ context.Context, _ *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	return &rpc.DiffResponse{}, nil
}

// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
func (p *xyzProvider) Create(_ context.Context, req *rpc.CreateRequest) (*rpc.CreateResponse, error) {

	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsMap := inputs.Mappable()

	typ := resource.URN(req.GetUrn()).Type()
	path := p.metadata.ResourceUrls[typ.String()]
	url := fmt.Sprintf("%s%s", p.metadata.BaseUrl, path)

	outputsMap, err := sendRequestWithTimeout("POST", url, inputsMap)
	if err != nil {
		return nil, err
	}

	// Note that we assume this particular structure of all resource IDs. The ID is then used for all
	// update, read, and delete operations.
	id := fmt.Sprintf("%s/%s", path, outputsMap["id"])

	outputs, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputsMap),
		plugin.MarshalOptions{SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &rpc.CreateResponse{
		Id:         id,
		Properties: outputs,
	}, nil
}

// Read the current live state associated with a resource.
func (p *xyzProvider) Read(_ context.Context, req *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	id := req.GetId()
	url := fmt.Sprintf("%s%s", p.metadata.BaseUrl, id)

	outputsMap, err := sendRequestWithTimeout("GET", url, nil)
	if err != nil {
		return nil, err
	}

	outputs, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputsMap),
		plugin.MarshalOptions{KeepSecrets: true, KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &rpc.ReadResponse{
		Id:         id,
		Properties: outputs,
		Inputs:     req.GetProperties(),
	}, nil
}

// Update updates an existing resource with new values.
func (p *xyzProvider) Update(_ context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	url := fmt.Sprintf("%s%s", p.metadata.BaseUrl, req.GetId())

	inputs, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsMap := inputs.Mappable()

	outputsMap, err := sendRequestWithTimeout("PATCH", url, inputsMap)
	if err != nil {
		return nil, err
	}

	outputs, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputsMap),
		plugin.MarshalOptions{SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &rpc.UpdateResponse{
		Properties: outputs,
	}, nil
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (p *xyzProvider) Delete(_ context.Context, req *rpc.DeleteRequest) (*pbempty.Empty, error) {
	url := fmt.Sprintf("%s%s", p.metadata.BaseUrl, req.GetId())

	_, err := sendRequestWithTimeout("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	return &pbempty.Empty{}, nil
}

// Construct creates a new component resource.
func (p *xyzProvider) Construct(_ context.Context, _ *rpc.ConstructRequest) (*rpc.ConstructResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Construct is not yet implemented")
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (p *xyzProvider) GetPluginInfo(context.Context, *pbempty.Empty) (*rpc.PluginInfo, error) {
	return &rpc.PluginInfo{
		Version: p.version,
	}, nil
}

// GetSchema returns the JSON-serialized schema for the provider.
func (p *xyzProvider) GetSchema(_ context.Context, _ *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetSchema is not yet implemented")
}

// Cancel signals the provider to gracefully shut down and abort any ongoing resource operations.
// Operations aborted in this way will return an error (e.g., `Update` and `Create` will either a
// creation error or an initialization error). Since Cancel is advisory and non-blocking, it is up
// to the host to decide how long to wait after Cancel is called before (e.g.)
// hard-closing any gRPC connection.
func (p *xyzProvider) Cancel(context.Context, *pbempty.Empty) (*pbempty.Empty, error) {
	return &pbempty.Empty{}, nil
}

func sendRequestWithTimeout(method, rawurl string, body map[string]interface{}) (map[string]interface{}, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")

	var res *http.Response
	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, rawurl, &buf)
	if err != nil {
		return nil, err
	}
	req.Header = reqHeaders

	client := http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 300 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		return nil, errors.Errorf("HTTP request failed with %v: %s", res.StatusCode, body)
	}

	if res.StatusCode == 204 {
		return nil, nil
	}

	result := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		return nil, errors.Wrapf(err, "decoding JSON %s", body)
	}

	return result, nil
}
