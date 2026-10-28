// Package listutil holds list helpers shared by the resources.
package listutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SameElements reports whether a and b hold the same elements in any order.
func SameElements[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[T]int, len(a))
	for _, v := range a {
		counts[v]++
	}
	for _, v := range b {
		if counts[v] == 0 {
			return false
		}
		counts[v]--
	}
	return true
}

// KeepOrder returns prior when it holds the same elements as got, so a
// server-side reordering does not show up as a diff.
func KeepOrder[T comparable](prior, got []T) []T {
	// An empty got is returned as is, so [] and null stay distinct.
	if len(got) > 0 && SameElements(prior, got) {
		return prior
	}
	return got
}

// Strings returns the list's elements as a non-nil slice, so an empty list
// is sent to the API as [] rather than omitted.
func Strings(ctx context.Context, list types.List) []string {
	out := []string{}
	list.ElementsAs(ctx, &out, false)
	return out
}
