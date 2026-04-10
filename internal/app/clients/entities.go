package clients

import "github.com/luikyv/go-oidc/pkg/goidc"

// Client is the safe client projection. Never includes the secret.
type Client struct {
	ID                 string
	RedirectURIs       []string
	GrantTypes         []string
	ResponseTypes      []string
	ScopeIDs           string
	TokenAuthnMethod   string
	CreatedAtTimestamp int
	ExpiresAtTimestamp int
}

// CreateRes is returned by [Create.Create]. Secret is the plaintext client
// secret — it is not stored in plaintext and cannot be recovered after this call.
type CreateRes struct {
	ID     string
	Secret string
}

func clientFromGoidc(c *goidc.Client) *Client {
	grantTypes := make([]string, len(c.GrantTypes))
	for i, gt := range c.GrantTypes {
		grantTypes[i] = string(gt)
	}
	responseTypes := make([]string, len(c.ResponseTypes))
	for i, rt := range c.ResponseTypes {
		responseTypes[i] = string(rt)
	}
	return &Client{
		ID:                 c.ID,
		RedirectURIs:       c.RedirectURIs,
		GrantTypes:         grantTypes,
		ResponseTypes:      responseTypes,
		ScopeIDs:           c.ScopeIDs,
		TokenAuthnMethod:   string(c.TokenAuthnMethod),
		CreatedAtTimestamp: c.CreatedAtTimestamp,
		ExpiresAtTimestamp: c.ExpiresAtTimestamp,
	}
}
