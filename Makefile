PROJECT_NAME := Pulumi Xyz Resource Provider

PACK            := xyz
PACKDIR         := sdk
PROJECT         := github.com/pulumi/pulumi-${PACK}
PROVIDER        := pulumi-resource-${PACK}
CODEGEN         := pulumi-sdkgen-${PACK}
VERSION         := $(shell pulumictl get version)

WORKING_DIR     := $(shell pwd)

VERSION_FLAGS   := -ldflags "-X github.com/pulumi/pulumi-${PACK}/provider/pkg/version.Version=${VERSION}"

ensure::
	@echo "GO111MODULE=on go mod download"; cd pkg; GO111MODULE=on go mod download

codegen::
	(cd pkg && go build -a -o $(WORKING_DIR)/bin/$(CODEGEN) $(VERSION_FLAGS) $(PROJECT)/cmd/$(CODEGEN))

provider::
	(cd pkg && go build -a -o $(WORKING_DIR)/bin/$(PROVIDER) $(VERSION_FLAGS) $(PROJECT)/cmd/$(PROVIDER))

generate::
	$(WORKING_DIR)/bin/$(CODEGEN) ${PACKDIR} ${VERSION}

build_nodejs:: VERSION := $(shell pulumictl get version --language javascript)
build_nodejs::
	cd ${PACKDIR}/nodejs/ && \
	yarn install && \
	node --max-old-space-size=4096 /usr/local/bin/tsc --diagnostics && \
	cp ../../README.md ../../LICENSE package.json yarn.lock ./bin/ && \
	sed -i.bak -e "s/\$${VERSION}/$(VERSION)/g" ./bin/package.json

build_python:: VERSION := $(shell pulumictl get version --language python)
build_python::
	cd sdk/python/ && \
	cp ../../README.md . && \
	python3 setup.py clean --all 2>/dev/null && \
	rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
	sed -i.bak -e "s/\$${VERSION}/$(VERSION)/g" -e "s/\$${PLUGIN_VERSION}/$(VERSION)/g" ./bin/setup.py && \
	rm ./bin/setup.py.bak && \
	cd ./bin && python3 setup.py build sdist

build_dotnet:: DOTNET_VERSION := $(shell pulumictl get version --language dotnet)
build_dotnet::
	cd ${PACKDIR}/dotnet/ && \
		echo "azure-native\n${DOTNET_VERSION}" >version.txt && \
		dotnet build /p:Version=${DOTNET_VERSION}

clean::
	rm -rf sdk/nodejs && mkdir sdk/nodejs && touch sdk/nodejs/go.mod
	rm -rf sdk/python && mkdir sdk/python && touch sdk/python/go.mod && cp README.md sdk/python
	rm -rf sdk/dotnet && mkdir sdk/dotnet && touch sdk/dotnet/go.mod
	rm -rf sdk/go/${PACK}

install_dotnet_sdk::
	mkdir -p $(WORKING_DIR)/nuget
	find . -name '*.nupkg' -print -exec cp -p {} ${WORKING_DIR}/nuget \;

install_python_sdk::

install_go_sdk::

install_nodejs_sdk::
	yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin

build:: clean codegen generate provider build_sdks install_sdks
build_sdks: build_nodejs build_dotnet build_python
install_sdks:: install_dotnet_sdk install_python_sdk install_nodejs_sdk

.PHONY: ensure generate build_provider build
