// Package writeonly handles arguments the API accepts but never returns, such
// as custom SSL keys. After an import they are null in state, so a configured
// value is recorded instead of forcing a replacement.
package writeonly

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

const importedKey = "imported"

type privateState interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

type privateGetter interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

// MarkImported records in private state that the resource was imported.
func MarkImported(ctx context.Context, private privateState) diag.Diagnostics {
	return private.SetKey(ctx, importedKey, []byte("true"))
}

// requiresReplace reports whether a change must replace the resource: always,
// unless the resource was imported and no value has been recorded yet.
func requiresReplace(ctx context.Context, private privateGetter, stateIsNull bool) (bool, diag.Diagnostics) {
	if !stateIsNull {
		return true, nil
	}
	imported, diags := private.GetKey(ctx, importedKey)
	return len(imported) == 0, diags
}

const description = "Changing the value creates a new resource, except after an import, when it is only recorded."

// RequiresReplace is the string variant of the replacement rule.
func RequiresReplace() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		replace, diags := requiresReplace(ctx, req.Private, req.StateValue.IsNull())
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = replace
	}, description, description)
}

// RequiresReplaceObject is the object variant of the replacement rule.
func RequiresReplaceObject() planmodifier.Object {
	return objectplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
		replace, diags := requiresReplace(ctx, req.Private, req.StateValue.IsNull())
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = replace
	}, description, description)
}
