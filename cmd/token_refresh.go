package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config keys login --web saves next to token for a Dex session, so the
// short-lived Dex token can be renewed. A PAT has none of them.
const (
	keyRefreshToken   = "refresh_token"
	keyTokenExpiresAt = "token_expires_at" // Unix seconds
	keyTokenEndpoint  = "token_endpoint"
	keyClientID       = "client_id"
)

var refreshKeys = []string{keyRefreshToken, keyTokenExpiresAt, keyTokenEndpoint, keyClientID}

// refreshMargin renews a saved token that expires within it, so it does
// not run out during the request.
const refreshMargin = time.Minute

var errSessionExpired = errors.New("session expired, run terrakube login --web")

// setSession writes tok to settings as the saved token. A response with a
// refresh_token (Dex) also stores what renewing it needs; one without (the
// Terrakube login broker) removes refresh data an earlier login left.
func setSession(settings map[string]any, tok *tokenResponse, tokenURL, clientID string) {
	settings["token"] = tok.AccessToken
	for _, k := range refreshKeys {
		delete(settings, k)
	}
	if tok.RefreshToken == "" {
		return
	}
	settings[keyRefreshToken] = tok.RefreshToken
	settings[keyTokenEndpoint] = tokenURL
	settings[keyClientID] = clientID
	if tok.ExpiresIn > 0 {
		settings[keyTokenExpiresAt] = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second).Unix()
	}
}

// refreshSavedToken renews the saved Dex token when it has expired or
// expires within refreshMargin: it exchanges the refresh_token at the saved
// token endpoint and saves the new tokens. A token from TERRAKUBE_TOKEN or
// TERRAKUBE_PAT is used as is, and a saved token without refresh data or
// expiry (a PAT, or a broker token) is never touched.
func refreshSavedToken() error {
	if os.Getenv("TERRAKUBE_TOKEN") != "" || os.Getenv("TERRAKUBE_PAT") != "" {
		return nil
	}
	refresh := viper.GetString(keyRefreshToken)
	expiresAt := viper.GetInt64(keyTokenExpiresAt)
	if refresh == "" || expiresAt == 0 || time.Until(time.Unix(expiresAt, 0)) > refreshMargin {
		return nil
	}
	tokenURL, clientID := viper.GetString(keyTokenEndpoint), viper.GetString(keyClientID)
	if tokenURL == "" {
		return errSessionExpired
	}
	tok, err := requestToken(getContext(), tokenURL, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refresh},
		"client_id":     {clientID},
	})
	if err != nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Refreshing the saved token:", err)
		}
		return errSessionExpired
	}
	if tok.RefreshToken == "" {
		tok.RefreshToken = refresh // not rotated: the old one stays valid
	}

	var saved map[string]any
	if err := updateConfig(func(settings map[string]any) {
		setSession(settings, tok, tokenURL, clientID)
		saved = settings
	}); err != nil {
		return fmt.Errorf("saving refreshed token: %w", err)
	}
	// viper read the config before the refresh; later reads in this run
	// (currentToken, another newClient) must see the new session.
	for _, k := range append([]string{"token"}, refreshKeys...) {
		v, ok := saved[k]
		if !ok {
			v = ""
		}
		viper.Set(k, v)
	}
	return nil
}
