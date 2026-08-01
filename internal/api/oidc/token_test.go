package oidc

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestNormalizeSingleAudienceClientIDs(t *testing.T) {
	clientIDs := normalizeSingleAudienceClientIDs([]string{
		" shopify-client ",
		"",
		"shopify-client",
		"\tportal-client\n",
	})

	if len(clientIDs) != 2 {
		t.Fatalf("normalizeSingleAudienceClientIDs() returned %d IDs, want 2", len(clientIDs))
	}
	for _, clientID := range []string{"shopify-client", "portal-client"} {
		if _, ok := clientIDs[clientID]; !ok {
			t.Errorf("normalizeSingleAudienceClientIDs() is missing %q", clientID)
		}
	}
}

func TestSelectIDTokenAudience(t *testing.T) {
	tests := []struct {
		name       string
		clientID   string
		audience   []string
		configured []string
		want       []string
	}{
		{
			name:       "allowlisted client",
			clientID:   "shopify-client",
			audience:   []string{"shopify-client", "project-id"},
			configured: []string{"shopify-client"},
			want:       []string{"shopify-client"},
		},
		{
			name:       "unlisted client",
			clientID:   "portal-client",
			audience:   []string{"portal-client", "shopify-client", "project-id"},
			configured: []string{"shopify-client"},
			want:       []string{"portal-client", "shopify-client", "project-id"},
		},
		{
			name:     "empty configuration",
			clientID: "shopify-client",
			audience: []string{"shopify-client", "project-id"},
			want:     []string{"shopify-client", "project-id"},
		},
		{
			name:       "unknown configured client",
			clientID:   "shopify-client",
			audience:   []string{"shopify-client", "project-id"},
			configured: []string{"unknown-client"},
			want:       []string{"shopify-client", "project-id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectIDTokenAudience(tt.clientID, tt.audience, normalizeSingleAudienceClientIDs(tt.configured))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("selectIDTokenAudience() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectIDTokenAudienceDoesNotMutateSessionAudience(t *testing.T) {
	audience := []string{"shopify-client", "project-id"}
	got := selectIDTokenAudience(
		"shopify-client",
		audience,
		normalizeSingleAudienceClientIDs([]string{"shopify-client"}),
	)

	got[0] = "changed"
	if audience[0] != "shopify-client" {
		t.Fatalf("selectIDTokenAudience() mutated the session audience: %v", audience)
	}
}

func TestSelectIDTokenAudienceProducesSingletonClaim(t *testing.T) {
	const clientID = "shopify-client"
	audience := selectIDTokenAudience(
		clientID,
		[]string{clientID, "project-id"},
		normalizeSingleAudienceClientIDs([]string{clientID}),
	)
	claims := oidc.NewIDTokenClaims(
		"https://issuer.example.com",
		"user-id",
		audience,
		time.Now().Add(time.Hour),
		time.Now(),
		"",
		"",
		nil,
		clientID,
		0,
	)

	encoded, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var token struct {
		Audience []string `json:"aud"`
	}
	if err := json.Unmarshal(encoded, &token); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !slices.Equal(token.Audience, []string{clientID}) {
		t.Fatalf("encoded ID-token audience = %v, want [%s]", token.Audience, clientID)
	}
}
