package client

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestUsedEndpoints fails when the client calls an endpoint that is missing
// from UsedEndpoints, or UsedEndpoints lists one the client no longer calls.
func TestUsedEndpoints(t *testing.T) {
	call := regexp.MustCompile(`(?:c\.do\(ctx, http\.Method(\w+), |Paginate\[\w+\]\(ctx, c, ())(?:path\()?(\w+\+)?"([^"]+)"`)
	placeholder := regexp.MustCompile(`\{[^}]+\}|%s`)
	norm := func(method, p string) string {
		return strings.ToUpper(method) + " " + placeholder.ReplaceAllString(p, "{}")
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	called := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range call.FindAllStringSubmatch(string(src), -1) {
			method := m[1]
			if method == "" {
				method = "GET"
			}
			if m[3] == "" {
				called[norm(method, m[4])] = true
				continue
			}
			// serviceType-prefixed paths are shared by several services.
			for _, svc := range []string{"/applications/{}", "/static-sites/{}", "/load-balancers/{}"} {
				called[norm(method, svc+strings.TrimPrefix(m[4], "/%s"))] = true
			}
		}
	}

	listed := map[string]bool{}
	for _, ep := range UsedEndpoints {
		listed[norm(ep.Method, strings.TrimPrefix(ep.Path, "/v3"))] = true
	}
	for ep := range listed {
		if !called[ep] {
			t.Errorf("UsedEndpoints lists %s, which the client never calls", ep)
		}
	}
	for ep := range called {
		// Only applications and static sites have env vars.
		if !listed[ep] && !strings.Contains(ep, "/load-balancers/{}/env-vars") {
			t.Errorf("client calls %s, which is missing from UsedEndpoints", ep)
		}
	}
}
