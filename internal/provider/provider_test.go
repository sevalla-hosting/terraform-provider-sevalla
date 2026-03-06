package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProviderSchema(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("unexpected error creating provider server: %v", err)
	}

	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("unexpected error getting provider schema: %v", err)
	}
	if resp.Diagnostics != nil {
		for _, d := range resp.Diagnostics {
			if d.Severity == tfprotov6.DiagnosticSeverityError {
				t.Errorf("unexpected error diagnostic: %s", d.Summary)
			}
		}
	}

	if resp.Provider == nil {
		t.Fatal("expected provider schema, got nil")
	}

	found := false
	for _, attr := range resp.Provider.Block.Attributes {
		if attr.Name == "api_key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected api_key attribute in provider schema")
	}
}
