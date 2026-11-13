package cmd

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

func TestResolveByName(t *testing.T) {
	fields := func(o *terrakube.Organization) (string, string) { return o.ID, o.Name }
	one := []*terrakube.Organization{{ID: "a", Name: "acme"}}
	two := []*terrakube.Organization{{ID: "a", Name: "acme"}, {ID: "b", Name: "acme"}}

	tests := []struct {
		name    string
		items   []*terrakube.Organization
		err     error
		lookup  string
		want    string
		wantErr string
	}{
		{name: "single match", items: one, lookup: "acme", want: "a"},
		{name: "not found", lookup: "acme", wantErr: `no organization found with name "acme"`},
		{name: "multiple", items: two, lookup: "acme", wantErr: `multiple organizations match name "acme", use --organization with the ID`},
		{name: "exact match among wildcard hits", items: []*terrakube.Organization{{ID: "x", Name: "acme-prod"}, {ID: "a", Name: "acme"}}, lookup: "acme", want: "a"},
		{name: "wildcard is not a match", items: []*terrakube.Organization{{ID: "p", Name: "production"}}, lookup: "prod*", wantErr: `no organization found with name "prod*"`},
		{name: "list error", err: errors.New("boom"), lookup: "acme", wantErr: "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveByName(tt.items, tt.err, "organization", tt.lookup, fields)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("expected %q, got %q (err %v)", tt.want, got, err)
			}
		})
	}
}

func TestOrgResolverQuotesName(t *testing.T) {
	resetGlobalFlags()

	var filter string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter = r.URL.Query().Get("filter")
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{{ID: "org-1", Name: `my "org" x\y`}})
	}))
	defer ts.Close()

	c, err := newClient()
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	got, err := orgResolver(context.Background(), c, nil, `my "org" x\y`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "org-1" {
		t.Errorf("expected org-1, got %q", got)
	}
	if want := `name=="my \"org\" x\\y"`; filter != want {
		t.Errorf("expected filter %s, got %s", want, filter)
	}
}
