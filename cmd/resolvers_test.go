package cmd

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
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

func TestNamedParentResolvers(t *testing.T) {
	const orgID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	tests := []struct {
		name       string
		resource   string // value passed to the parent flag
		args       []string
		segment    string // list path segment under the organization
		filterKey  string // query parameter carrying the RSQL filter
		resolvedID string
		payload    any
		followPath string // substring expected in the follow-up request path
	}{
		{
			name:       "collection",
			resource:   "my coll",
			args:       []string{"collection-item", "list", "--collection", "my coll"},
			segment:    "collection",
			filterKey:  "filter[collection]",
			resolvedID: "coll-1",
			payload:    []*terrakube.Collection{{ID: "coll-1", Name: "my coll"}},
			followPath: "/collection/coll-1/item",
		},
		{
			name:       "provider",
			resource:   "aws",
			args:       []string{"provider-version", "list", "--provider", "aws"},
			segment:    "provider",
			filterKey:  "filter[provider]",
			resolvedID: "prov-1",
			payload:    []*terrakube.Provider{{ID: "prov-1", Name: "aws"}},
			followPath: "/provider/prov-1/version",
		},
		{
			name:       "module",
			resource:   "vpc",
			args:       []string{"module-version", "list", "--module", "vpc"},
			segment:    "module",
			filterKey:  "filter[module]",
			resolvedID: "mod-1",
			payload:    []*terrakube.Module{{ID: "mod-1", Name: "vpc"}},
			followPath: "/module/mod-1/",
		},
		{
			name:       "notification-configuration",
			resource:   "slack",
			args:       []string{"notification-trigger", "list", "--notification-configuration", "slack"},
			segment:    "notificationConfiguration",
			filterKey:  "filter[notification_configuration]",
			resolvedID: "nc-1",
			payload:    []*terrakube.NotificationConfiguration{{ID: "nc-1", Name: "slack"}},
			followPath: "/notificationConfiguration/nc-1/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			var filters, paths []string
			ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				w.Header().Set("Content-Type", "application/vnd.api+json")
				if strings.HasSuffix(r.URL.Path, "/organization/"+orgID+"/"+tt.segment) {
					filters = append(filters, r.URL.Query().Get(tt.filterKey))
					_ = jsonapi.MarshalPayload(w, tt.payload)
					return
				}
				_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{})
			}))
			defer ts.Close()

			args := slices.Concat(tt.args, []string{"--organization-id", orgID})
			if _, err := executeCommand(args...); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(filters) != 1 {
				t.Fatalf("expected one resolver request, got %v (paths %v)", filters, paths)
			}
			if want := `name=="` + tt.resource + `"`; filters[0] != want {
				t.Errorf("filter = %q, want %q", filters[0], want)
			}
			last := paths[len(paths)-1]
			if !strings.Contains(last, tt.followPath) {
				t.Errorf("follow-up path %q does not contain resolved ID path %q", last, tt.followPath)
			}
		})
	}
}

func TestNamedParentResolverNotFound(t *testing.T) {
	resetGlobalFlags()
	var requests int
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, []*terrakube.Provider{})
	}))
	defer ts.Close()

	_, err := executeCommand("provider-version", "list", "--provider", "ghost",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err == nil || !strings.Contains(err.Error(), `no provider found with name "ghost"`) {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if requests != 1 {
		t.Errorf("expected only the resolver request, got %d requests", requests)
	}
}
