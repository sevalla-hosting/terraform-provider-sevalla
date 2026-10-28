package application

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// After an import the certificate is only recorded; the user must be told.
func TestApplicationDomainUpdate_WarnsWhenCertificateIsOnlyRecorded(t *testing.T) {
	ctx := context.Background()
	r := NewDomainResource()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema

	state := ApplicationDomainResourceModel{
		ID: types.StringValue("dom-1"), ApplicationID: types.StringValue("app-1"), Name: types.StringValue("shop.example.com"),
		Type: types.StringValue("CUSTOM"), IsPrimary: types.BoolValue(true), IsWildcard: types.BoolValue(false),
		IsEnabled: types.BoolValue(true), Status: types.StringValue("ACTIVE"), DNSRecords: types.StringNull(),
		CustomSSLCert: types.StringNull(), CustomSSLKey: types.StringNull(),
		CreatedAt: types.StringValue("2025-01-01T00:00:00Z"), UpdatedAt: types.StringValue("2025-01-01T00:00:00Z"),
	}
	plan := state
	plan.CustomSSLCert = types.StringValue("CERT")
	plan.CustomSSLKey = types.StringValue("KEY")

	newState := func(model ApplicationDomainResourceModel) (tfsdk.State, diag.Diagnostics) {
		st := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
		return st, st.Set(ctx, &model)
	}
	stateValue, diags := newState(state)
	planValue, d := newState(plan)
	diags.Append(d...)
	if diags.HasError() {
		t.Fatalf("building request: %v", diags)
	}

	req := resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: s, Raw: planValue.Raw},
		State: stateValue,
	}
	resp := resource.UpdateResponse{State: tfsdk.State{Schema: s, Raw: stateValue.Raw}}
	r.Update(ctx, req, &resp)

	if resp.Diagnostics.HasError() || resp.Diagnostics.WarningsCount() != 1 {
		t.Fatalf("expected exactly one warning, got %v", resp.Diagnostics)
	}
	var got ApplicationDomainResourceModel
	resp.State.Get(ctx, &got)
	if got.CustomSSLCert.ValueString() != "CERT" {
		t.Errorf("expected the certificate to be recorded, got %s", got.CustomSSLCert)
	}
}
