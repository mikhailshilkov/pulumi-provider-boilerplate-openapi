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

package gen

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-xyz/pkg/provider"
	"github.com/pulumi/pulumi/pkg/v3/codegen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Schema builds the Pulumi schema from an Open API spec. It also returns extra metadata that is not included in
// the schema but is crucial for the provider at runtime (e.g., API endpoints).
func Schema() (*pschema.PackageSpec, *provider.APIMetadata, error) {
	swagger, err := loadSwaggerSpec()
	if err != nil {
		return nil, nil, err
	}

	pkg := pschema.PackageSpec{
		Name: "xyz",
		Language: map[string]json.RawMessage{
			"nodejs": rawMessage(map[string]interface{}{
				"dependencies": map[string]string{
					"@pulumi/pulumi": "^3.0.0",
				},
			}),
			"python": rawMessage(map[string]interface{}{
				"usesIOClasses": true,
			}),
			"csharp": rawMessage(map[string]interface{}{
				"packageReferences": map[string]string{
					"Pulumi":                       "3.*",
					"System.Collections.Immutable": "1.6.0",
				},
			}),
			"go": rawMessage(map[string]interface{}{}),
		},
		Types:     map[string]pschema.ComplexTypeSpec{},
		Resources: map[string]pschema.ResourceSpec{},
		Functions: map[string]pschema.FunctionSpec{},
	}
	metadata := provider.APIMetadata{
		BaseUrl:      fmt.Sprintf("%s://%s%s", swagger.Schemes[0], swagger.Host, swagger.BasePath),
		ResourceUrls: map[string]string{},
	}

	// Discover all API paths and build a map of resources and resource operations.
	resourceMap := map[string]map[string]*spec.Operation{}
	for path, pathItem := range swagger.Paths.Paths {
		// We expect POST, GET, PATCH, and DELETE to be present for each resource.
		// You may need to adjust this for the resource model of your API.
		for idx, op := range []*spec.Operation{pathItem.Post, pathItem.Get, pathItem.Patch, pathItem.Delete} {
			if op == nil {
				continue
			}

			// Operation ID is expected to be of the shape `Resource_Action`.
			parts := strings.Split(op.OperationProps.ID, "_")
			if len(parts) != 2 {
				continue
			}

			tok := fmt.Sprintf("%s:index:%s", pkg.Name, parts[0])
			action := parts[1]
			if v, ok := resourceMap[tok]; ok {
				v[action] = op
			} else {
				resourceMap[tok] = map[string]*spec.Operation{
					action: op,
				}
			}

			if idx == 0 {
				// Populate the resource creation URL for the provider metadata.
				metadata.ResourceUrls[tok] = path
			}
		}
	}

	g := packageGenerator{pkg: &pkg, swagger: swagger}
	for tok, res := range resourceMap {
		create, hasCreate := res["Create"]
		get, hasGet := res["Get"]
		_, hasUpdate := res["Update"]
		_, hasDelete := res["Delete"]
		if hasCreate && hasGet && hasUpdate && hasDelete {
			err = g.genResources(tok, create, get)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return &pkg, &metadata, nil
}

type packageGenerator struct {
	pkg     *pschema.PackageSpec
	swagger *spec.Swagger
}

func (g *packageGenerator) genResources(tok string, create, get *spec.Operation) error {
	resourceRequest, err := g.getBodyProperties(create.Parameters)
	if err != nil {
		return errors.Wrapf(err, "failed to generate '%s': request type", tok)
	}

	response, err := g.getResponseProperties(get.Responses.StatusCodeResponses)
	if err != nil {
		return errors.Wrapf(err, "failed to generate '%s': request type", tok)
	}

	resourceSpec := pschema.ResourceSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Type:       "object",
			Properties: response.props,
			Required:   response.required.SortedValues(),
		},
		InputProperties: resourceRequest.props,
		RequiredInputs:  resourceRequest.required.SortedValues(),
	}
	g.pkg.Resources[tok] = resourceSpec
	return nil
}

type bag struct {
	props    map[string]pschema.PropertySpec
	required codegen.StringSet
}

func (g *packageGenerator) getBodyProperties(parameters []spec.Parameter) (*bag, error) {
	for _, param := range parameters {
		switch {
		case param.In == "body":
			ptr := param.Schema.Ref.GetPointer()
			if ptr == nil || ptr.IsEmpty() {
				return nil, errors.New("expected a pointer in the schema")
			}

			value, _, err := ptr.Get(g.swagger)
			if err != nil {
				return nil, errors.Wrapf(err, "get pointer")
			}
			schema := value.(spec.Schema)

			return g.genProperties(&schema, false)
		default:
			return nil, errors.New("non-body parameters aren't supported for Create methods")
		}
	}

	return &bag{}, nil
}

func (g *packageGenerator) getResponseProperties(statusCodeResponses map[int]spec.Response) (*bag, error) {
	var codes []int
	for code := range statusCodeResponses {
		if code >= 300 || code < 200 {
			continue
		}

		codes = append(codes, code)
	}
	sort.Ints(codes)

	if len(codes) == 0 {
		return nil, errors.New("no 2xx response found")
	}

	// Find the lowest 2xx response with a schema definition and derive response properties from it.
	resp := statusCodeResponses[codes[0]]
	ptr := resp.Schema.Ref.GetPointer()
	if ptr == nil || ptr.IsEmpty() {
		return nil, errors.New("expected a pointer in the schema")
	}

	value, _, err := ptr.Get(g.swagger)
	if err != nil {
		return nil, errors.Wrapf(err, "get pointer")
	}
	schema := value.(spec.Schema)

	return g.genProperties(&schema, true /*isOutput*/)
}

func (g *packageGenerator) genProperties(schema *spec.Schema, isOutput bool) (*bag, error) {
	result := bag{
		props:    map[string]pschema.PropertySpec{},
		required: codegen.NewStringSet(schema.Required...),
	}

	for name, property := range schema.Properties {
		if !isOutput && property.ReadOnly {
			// Skip read-only properties for input types.
			continue
		}
		if isOutput && name == "id" {
			// Every Pulumi resource has an output called ID already, no need to add it to the schema.
			continue
		}

		var primitiveTypeName string
		if len(property.Type) > 0 {
			primitiveTypeName = property.Type[0]
		}
		propertySpec := pschema.PropertySpec{
			Description: property.Description,
			TypeSpec:    pschema.TypeSpec{Type: primitiveTypeName},
		}
		result.props[name] = propertySpec

		if isOutput {
			result.required.Add(name)
		}
	}

	return &result, nil
}

func loadSwaggerSpec() (*spec.Swagger, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "open-api-spec/todo-backend.json")

	bytes, err := swag.LoadFromFileOrHTTP(path)
	if err != nil {
		return nil, err
	}
	swagger := spec.Swagger{}
	err = swagger.UnmarshalJSON(bytes)
	if err != nil {
		return nil, err
	}

	return &swagger, nil
}

func rawMessage(v interface{}) json.RawMessage {
	bytes, err := json.Marshal(v)
	contract.Assert(err == nil)
	return bytes
}
