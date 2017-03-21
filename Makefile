REGISTRY := eu.gcr.io/jetstack-sandbox
IMAGE_NAME := lieutenant-elastic-search
BUILD_TAG := build
IMAGE_TAGS := canary

all: build docker_build

build:
	cd sidecar; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o sidecar_linux_amd64 .

docker_build:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) .

docker_push: docker_build
	set -e; \
		for tag in $(IMAGE_TAGS); do \
		docker tag $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) $(REGISTRY)/$(IMAGE_NAME):$${tag} ; \
		docker push $(REGISTRY)/$(IMAGE_NAME):$${tag}; \
	done

test: test_helm test_golang

test_helm:
	helm lint contrib/charts/elasticsearch

test_golang:
	go test $(shell go list ./sidecar/... | grep -v '/vendor/')

fmt_golang:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please make sure you run go fmt!"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

vet_golang:
	go vet $(shell go list ./sidecar/... | grep -v '/vendor/')
