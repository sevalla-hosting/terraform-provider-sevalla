package application

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Unknown values come from Optional+Computed attributes that are not
// configured; they must never be sent as "" or 0.
func TestBuildUpdateProcessRequest_SkipsUnknownValues(t *testing.T) {
	state := ApplicationProcessResourceModel{
		DisplayName:      types.StringValue("web"),
		Entrypoint:       types.StringValue("npm start"),
		Port:             types.Int64Value(8080),
		IsIngressEnabled: types.BoolValue(true),
		IngressProtocol:  types.StringValue("http"),
		TimeZone:         types.StringValue("UTC"),
		ScalingStrategy:  types.ObjectNull(scalingStrategyAttrTypes),
		LivenessProbe:    types.ObjectNull(probeAttrTypes),
		ReadinessProbe:   types.ObjectNull(probeAttrTypes),
	}
	plan := state
	plan.DisplayName = types.StringValue("web-renamed")
	plan.Entrypoint = types.StringUnknown()
	plan.Port = types.Int64Unknown()
	plan.IsIngressEnabled = types.BoolUnknown()
	plan.IngressProtocol = types.StringUnknown()
	plan.TimeZone = types.StringUnknown()
	plan.LivenessProbe = types.ObjectUnknown(probeAttrTypes)

	req, diags := buildUpdateProcessRequest(context.Background(), &plan, &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.DisplayName == nil || *req.DisplayName != "web-renamed" {
		t.Errorf("expected display_name to be sent, got %v", req.DisplayName)
	}
	if req.Entrypoint != nil || req.Port != nil || req.IsIngressEnabled != nil || req.IngressProtocol != nil || req.TimeZone != nil || req.LivenessProbe != nil {
		t.Errorf("unknown values were sent: %+v", req)
	}
}

func TestBuildUpdateProcessRequest_ClearsRemovedSchedule(t *testing.T) {
	state := ApplicationProcessResourceModel{
		Schedule:              types.StringValue("0 * * * *"),
		ActiveDeadlineSeconds: types.Int64Value(600),
		ScalingStrategy:       types.ObjectNull(scalingStrategyAttrTypes),
		LivenessProbe:         types.ObjectNull(probeAttrTypes),
		ReadinessProbe:        types.ObjectNull(probeAttrTypes),
	}
	plan := state
	plan.Schedule = types.StringNull()
	plan.ActiveDeadlineSeconds = types.Int64Null()

	req, _ := buildUpdateProcessRequest(context.Background(), &plan, &state)
	if req.Schedule == nil || *req.Schedule != "" {
		t.Errorf("expected schedule to be cleared with an empty string, got %v", req.Schedule)
	}
	if req.ActiveDeadlineSeconds == nil || *req.ActiveDeadlineSeconds != 0 {
		t.Errorf("expected active_deadline_seconds to be cleared with 0, got %v", req.ActiveDeadlineSeconds)
	}
}
