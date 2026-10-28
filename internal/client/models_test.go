package client

import (
	"encoding/json"
	"testing"
)

func TestNullable_Marshal(t *testing.T) {
	v := "p1"
	for _, tc := range []struct {
		in   Nullable[string]
		want string
	}{
		{Nullable[string]{}, `{}`},
		{Null[string](), `{"project_id":null}`},
		{NullableOf(&v), `{"project_id":"p1"}`},
		{NullableOf[string](nil), `{"project_id":null}`},
	} {
		b, err := json.Marshal(struct {
			ProjectID Nullable[string] `json:"project_id,omitzero"`
		}{tc.in})
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != tc.want {
			t.Errorf("got %s, want %s", b, tc.want)
		}
	}
}
