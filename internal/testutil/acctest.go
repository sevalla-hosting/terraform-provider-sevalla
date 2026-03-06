package testutil

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/provider"
)

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sevalla": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("SEVALLA_API_KEY"); v == "" {
		t.Fatal("SEVALLA_API_KEY must be set for acceptance tests")
	}
}
