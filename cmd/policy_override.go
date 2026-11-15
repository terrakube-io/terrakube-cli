package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.PolicyOverride]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "policy-override",
		Aliases: []string{"policy-overrides"},
		Parents: append(jobScope[:2:2], resource.ParentScope{
			Name:   "policy-evaluation",
			Flag:   "policy-evaluation",
			IDFlag: "policy-evaluation-id",
		}),
		Fields: []resource.FieldDef{
			{StructField: "Justification", Flag: "justification", Type: resource.String, Required: true, Description: "Why the soft-mandatory failure is accepted"},
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.PolicyOverride, error) {
			return c.PolicyOverrides.Get(ctx, pIDs[0], pIDs[1], pIDs[2], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.PolicyOverride, error) {
			// Who overrode and when are required attributes of an override.
			p.Attributes["overriddenAt"] = time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
			p.Attributes["overriddenBy"] = tokenUser(currentToken())
			return c.PolicyOverrides.CreateFields(ctx, pIDs[0], pIDs[1], pIDs[2], p)
		},
	})
}

// tokenUser returns the user a JWT names, read the way the server names the
// user of an override: email, then preferred_username, then name. It returns
// "cli-user", the server's own fallback, when the token carries none of them.
// The token is not verified: the server does that on the request itself.
func tokenUser(token string) string {
	var claims map[string]any
	if parts := strings.Split(token, "."); len(parts) == 3 {
		if payload, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
			_ = json.Unmarshal(payload, &claims)
		}
	}
	for _, key := range []string{"email", "preferred_username", "name"} {
		if v, _ := claims[key].(string); v != "" {
			return v
		}
	}
	return "cli-user"
}
