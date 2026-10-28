package listutil

import (
	"slices"
	"testing"
)

func TestKeepOrder(t *testing.T) {
	prior := []string{"b", "a", "a"}
	if got := KeepOrder(prior, []string{"a", "b", "a"}); !slices.Equal(got, prior) {
		t.Errorf("expected the prior order for the same elements, got %v", got)
	}
	changed := []string{"a", "b", "c"}
	if got := KeepOrder(prior, changed); !slices.Equal(got, changed) {
		t.Errorf("expected the new list when elements differ, got %v", got)
	}
	if got := KeepOrder([]string{"a", "a", "b"}, []string{"a", "b", "b"}); !slices.Equal(got, []string{"a", "b", "b"}) {
		t.Errorf("expected duplicates to be counted, got %v", got)
	}
	if got := KeepOrder(nil, []string{}); got == nil {
		t.Error("expected an empty API list to stay empty, not nil")
	}
}
