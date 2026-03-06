package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

type OpenAPISpec struct {
	Paths map[string]map[string]interface{} `json:"paths"`
}

func main() {
	specFile := "openapi.json"
	if len(os.Args) > 1 {
		specFile = os.Args[1]
	}

	data, err := os.ReadFile(specFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	hasBreaking := false

	for _, ep := range client.UsedEndpoints {
		pathItem, ok := spec.Paths[ep.Path]
		if !ok {
			fmt.Printf("BREAKING: endpoint removed: %s %s\n", ep.Method, ep.Path)
			hasBreaking = true
			continue
		}

		methodLower := methodToLower(ep.Method)
		if _, ok := pathItem[methodLower]; !ok {
			fmt.Printf("BREAKING: method removed: %s %s\n", ep.Method, ep.Path)
			hasBreaking = true
			continue
		}

		fmt.Printf("OK: %s %s\n", ep.Method, ep.Path)
	}

	if hasBreaking {
		fmt.Println("\nBREAKING changes detected!")
		os.Exit(1)
	}

	fmt.Println("\nAll endpoints OK.")
}

func methodToLower(method string) string {
	switch method {
	case "GET":
		return "get"
	case "POST":
		return "post"
	case "PUT":
		return "put"
	case "PATCH":
		return "patch"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}
