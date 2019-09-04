GOHOSTOS ?= linux
DOCKER_ARCHS ?= amd64
include Makefile.common
L1 := msaas
L2 := pasi
GIT_REPO := client

# use semantic versioning
# add more logic to automate: either replace or update version
THIS_PACKAGE_VERSION ?= 0.0.1
DOCKER_IMAGE_NAME       ?= $(GIT_REPO)

.PHONY: build
build:
	@echo ">> building binaries"
	env CGO_ENABLED=0 GOOS=$(GOHOSTOS) GOARCH=$(GOARCH) $(GO) build -a -installsuffix cgo -o $(THIS_PACKAGE_BINARY)/client .

.PHONY: run
run:
	@echo ">> running docker image"
	docker run -it $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$(DOCKER_ARCHS):$(DOCKER_IMAGE_TAG)