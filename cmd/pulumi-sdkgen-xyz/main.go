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

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-xyz/pkg/gen"
	"github.com/pulumi/pulumi-xyz/pkg/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tools"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"os"
	"path"

	"github.com/pkg/errors"
	dotnetgen "github.com/pulumi/pulumi/pkg/v3/codegen/dotnet"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsgen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pygen "github.com/pulumi/pulumi/pkg/v3/codegen/python"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: pulumi-sdkgen-xyz <target-sdk-folder> <version>\n")
		return
	}

	targetSdkFolder := os.Args[1]
	version := os.Args[2]

	err := emitPackage(targetSdkFolder, version)
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
	}
}

// emitPackage emits an entire package pack into the configured output directory with the configured settings.
func emitPackage(targetSdkFolder, version string) error {
	spec, metadata, err := gen.Schema()
	if err != nil {
		return errors.Wrap(err, "generating schema")
	}

	outdir := path.Join(".", "cmd", "pulumi-resource-xyz")
	err = emitSchema(spec, version, outdir, "main")
	if err != nil {
		return errors.Wrap(err, "writing schema")
	}
	err = emitMetadata(metadata, outdir, "main")
	if err != nil {
		return errors.Wrap(err, "writing schema")
	}

	ppkg, err := pschema.ImportSpec(*spec, nil)
	if err != nil {
		return errors.Wrap(err, "importing schema")
	}

	toolDescription := "the Pulumi SDK Generator"
	extraFiles := map[string][]byte{}

	sdkGenerators := map[string]func() (map[string][]byte, error){
		"python": func() (map[string][]byte, error) {
			return pygen.GeneratePackage(toolDescription, ppkg, extraFiles)
		},
		"nodejs": func() (map[string][]byte, error) {
			return nodejsgen.GeneratePackage(toolDescription, ppkg, extraFiles)
		},
		"go": func() (map[string][]byte, error) {
			return gogen.GeneratePackage(toolDescription, ppkg)
		},
		"dotnet": func() (map[string][]byte, error) {
			return dotnetgen.GeneratePackage(toolDescription, ppkg, extraFiles)
		},
	}

	for sdkName, generator := range sdkGenerators {
		files, err := generator()
		if err != nil {
			return errors.Wrapf(err, "generating %s package", sdkName)
		}

		for f, contents := range files {
			if err := emitFile(path.Join(targetSdkFolder, sdkName), f, contents); err != nil {
				return errors.Wrapf(err, "emitting file %v", f)
			}
		}
	}

	return nil
}

// emitSchema writes the Pulumi schema JSON to the 'schema.json' file in the given directory.
func emitSchema(pkgSpec *pschema.PackageSpec, version, outDir string, goPackageName string) error {
	schemaJSON, err := json.MarshalIndent(pkgSpec, "", "    ")
	if err != nil {
		return errors.Wrap(err, "marshaling Pulumi schema")
	}

	// Ensure the spec is stamped with a version.
	pkgSpec.Version = version

	compressedSchema := bytes.Buffer{}
	compressedWriter := gzip.NewWriter(&compressedSchema)
	err = json.NewEncoder(compressedWriter).Encode(pkgSpec)
	if err != nil {
		return errors.Wrap(err, "marshaling metadata")
	}
	if err = compressedWriter.Close(); err != nil {
		return err
	}

	err = emitFile(outDir, "schema.go", []byte(fmt.Sprintf(`package %s
var pulumiSchema = %#v
`, goPackageName, compressedSchema.Bytes())))
	if err != nil {
		return errors.Wrap(err, "saving metadata")
	}

	return emitFile(outDir, "schema.json", schemaJSON)
}

// emitMetadata writes the Metadata JSON to the 'metadata.json' file in the given directory.
func emitMetadata(metadata *provider.APIMetadata, outDir, goPackageName string) error {
	compressedMeta := bytes.Buffer{}
	compressedWriter := gzip.NewWriter(&compressedMeta)
	err := json.NewEncoder(compressedWriter).Encode(metadata)
	if err != nil {
		return errors.Wrap(err, "marshaling metadata")
	}

	if err = compressedWriter.Close(); err != nil {
		return err
	}

	formatted, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		return errors.Wrap(err, "marshaling metadata")
	}

	err = emitFile(outDir, "metadata.go", []byte(fmt.Sprintf(`package %s
var apiResources = %#v
`, goPackageName, compressedMeta.Bytes())))
	if err != nil {
		return err
	}

	return emitFile(outDir, "metadata.json", formatted)
}

func emitFile(outDir, relPath string, contents []byte) error {
	p := path.Join(outDir, relPath)
	if err := tools.EnsureDir(path.Dir(p)); err != nil {
		return errors.Wrap(err, "creating directory")
	}

	f, err := os.Create(p)
	if err != nil {
		return errors.Wrap(err, "creating file")
	}
	defer contract.IgnoreClose(f)

	_, err = f.Write(contents)
	return err
}
