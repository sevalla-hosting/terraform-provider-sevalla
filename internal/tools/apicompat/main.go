package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

type OpenAPISpec struct {
	Servers []struct {
		URL string `json:"url"`
	} `json:"servers"`
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

	basePath := serverBasePath(spec)

	hasBreaking := false

	for _, ep := range client.UsedEndpoints {
		lookup := stripBase(ep.Path, basePath)
		pathItem, ok := spec.Paths[lookup]
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

// serverBasePath extracts the base path (e.g. "/v3") from servers[0].url so
// that endpoint paths registered as "/v3/foo" match the spec's "/foo" keys.
func serverBasePath(spec OpenAPISpec) string {
	if len(spec.Servers) == 0 {
		return ""
	}
	raw := spec.Servers[0].URL
	if raw == "" {
		return ""
	}
	if u, err := url.Parse(raw); err == nil && u.Path != "" {
		return strings.TrimRight(u.Path, "/")
	}
	if i := strings.Index(raw, "://"); i >= 0 {
		rest := raw[i+3:]
		if j := strings.Index(rest, "/"); j >= 0 {
			return strings.TrimRight(rest[j:], "/")
		}
		return ""
	}
	return strings.TrimRight(raw, "/")
}

func stripBase(path, base string) string {
	if base == "" {
		return path
	}
	if strings.HasPrefix(path, base+"/") || path == base {
		return strings.TrimPrefix(path, base)
	}
	return path
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
