default: build

build:
	go build -o terraform-provider-sevalla

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sevalla-hosting/sevalla/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-sevalla ~/.terraform.d/plugins/registry.terraform.io/sevalla-hosting/sevalla/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

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
