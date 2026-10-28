default: build

# The local install must satisfy the version constraint in the configuration.
VERSION ?= $(shell git describe --tags --abbrev=0 | sed 's/^v//')

build:
	go build -o terraform-provider-sevalla

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sevalla/sevalla/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-sevalla ~/.terraform.d/plugins/registry.terraform.io/sevalla/sevalla/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./... -v -count=1

testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 120m

generate:
	go generate ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name sevalla --rendered-provider-name "Sevalla"

.PHONY: build install test testacc generate lint fmt docs
