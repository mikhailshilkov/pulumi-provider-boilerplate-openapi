# xyz Pulumi Open API-based Provider

This repo is a boilerplate showing how to create a native Pulumi provider. The resource model is generated automatically from a simple Open API specification.

You can search-replace `xyz` with the name of your desired provider as a starting point for creating a provider that manages resources in the target cloud.

## Navigate the repository

### Open API spec

A sample Open API (Swagger) specification is location in `open-api-spec/todo-backend.json`. It is based on the [`Todo-Backend`](https://www.todobackend.com/) project. The specification contains create/update/get/delete operations for a single resource: a `Todo`.

Please note that the sample specification is very simple and doesn't utilize a lot of more advanced features of Open API. The generation code is coupled to this particular specificaion and will likely not work for an arbitrary specification of your choice. All APIs are different and you will have to do the work of mapping your API to Pulumi resource model.

### Provider

Pulumi providers implement a gRPC protocol to connect to the Pulumi engine (CLI). The code for the provider implementation is in `pkg/provider/provider.go`. You will likely need to adjust this implementation to implement the features of your target API, including authentication, URL structures, parameter structure, response codes, error handling, and more.

### Code generator

A code generator is available which generates SDKs in TypeScript, Python, Go and .NET which are also checked in to the `sdk` folder. The SDKs are generated from the schema definitions based on the Open API spec described above.

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
