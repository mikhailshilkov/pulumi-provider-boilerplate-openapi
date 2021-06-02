# xyz Pulumi Provider

This repo is a boilerplate showing how to create a native Pulumi provider.  You can search-replace `xyz` with the name of your desired provider as a starting point for creating a provider that manages resources in the target cloud.

## Navigate the repository

### Resources

Custom resources are defined in `pkg/resources`. There is a separate Go file for each resource. The `pkg/resources/resource.go` file defines the "registry" of all resources registered within the provider.

The boilerplate repository comes with a single resource `RandomString` that generates a persistent random value of a given length. Try adding a new resource next to it while learning how the providers work.

### Provider gRPC

Pulumi providers implement a gRPC protocol to connect to the Pulumi engine (CLI). Most of the code for the provider implementation is in `pkg/provider/provider.go`. You shouldn't need to change this file for simple resources.

### Code generator

A code generator is available which generates SDKs in TypeScript, Python, Go and .NET which are also checked in to the `sdk` folder. The SDKs are generated from the schema definitions of the resources described above. Be sure to keep the `Schema` properties in sync with the resource CRUD operations.

### Example

An example of using the single resource defined in this example is in `examples/simple`.

## Build and test

```bash
# install the dependencies
make ensure

# build codegen, generate and build SDKs, build the provider
make build

# add the provider binary somewhere on your PATH
cp ./bin/pulumi-resource-xyz ~/go/bin

# test
$ cd examples/simple
$ yarn link @pulumi/xyz
$ pulumi stack init test
$ pulumi up
```

Note that the generated provider plugin (`pulumi-resource-xyz`) must be on your `PATH` to be used by Pulumi deployments.  If creating a provider for distribution to other users, you should ensure they install this plugin to their `PATH`.

## References

Other resources for learning about the Pulumi resource model:
* [Pulumi Kubernetes provider](https://github.com/pulumi/pulumi-kubernetes/blob/master/provider/pkg/provider/provider.go)
* [Pulumi Terraform Remote State provider](https://github.com/pulumi/pulumi-terraform/blob/master/provider/cmd/pulumi-resource-terraform/provider.go)
* [Dynamic Providers](https://www.pulumi.com/docs/intro/concepts/programming-model/#dynamicproviders)
