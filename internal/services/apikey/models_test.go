package apikey

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

func TestFlattenAPIKey(t *testing.T) {
	ctx := context.Background()
	resourceID := "app-1"
	key := &client.APIKey{
		ID:        "key-1",
		Name:      "ci",
		Enabled:   true,
		ExpiresAt: stringPtr("2027-01-01T00:00:00.000Z"),
		Roles: []client.APIKeyRoleRef{
			{ID: "role-b", Name: "B"},
			{ID: "role-a", Name: "A"},
		},
		Capabilities: []client.APIKeyCapability{{Permission: "APP:READ", ResourceID: &resourceID}},
	}

	roleIDs, _ := types.ListValueFrom(ctx, types.StringType, []string{"role-a", "role-b"})
	model := APIKeyResourceModel{
		ExpiresAt:    types.StringValue("2027-01-01T00:00:00Z"),
		RoleIDs:      roleIDs,
		Capabilities: types.ListNull(capabilityObjectType),
		Token:        types.StringValue("secret"),
	}
	if diags := flattenAPIKey(ctx, key, &model); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ExpiresAt.ValueString() != "2027-01-01T00:00:00Z" {
		t.Errorf("expected the configured timestamp to be kept, got %s", model.ExpiresAt)
	}
	if !model.RoleIDs.Equal(roleIDs) {
		t.Errorf("expected role_ids in configured order, got %s", model.RoleIDs)
	}
	var caps []APIKeyCapabilityModel
	model.Capabilities.ElementsAs(ctx, &caps, false)
	if len(caps) != 1 || caps[0].IDResource.ValueString() != "app-1" {
		t.Errorf("expected the capability's resource_id to be read, got %v", caps)
	}
	if model.Token.ValueString() != "secret" {
		t.Errorf("expected the token to be kept, got %s", model.Token)
	}

	// An empty role list is null, so configs that only use capabilities have no diff.
	key.Roles = nil
	if diags := flattenAPIKey(ctx, key, &model); diags.HasError() || !model.RoleIDs.IsNull() {
		t.Errorf("expected null role_ids, got %s (%v)", model.RoleIDs, diags)
	}
}

func stringPtr(s string) *string { return &s }
