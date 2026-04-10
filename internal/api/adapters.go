package api

import (
	"net/url"

	appclients "github.com/4nd3r5on/oidc-serv/internal/app/clients"
	appusers "github.com/4nd3r5on/oidc-serv/internal/app/users"
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
)

func toOIDCClient(c *appclients.Client) *genapi.OIDCClient {
	redirectURIs := make([]url.URL, 0, len(c.RedirectURIs))
	for _, raw := range c.RedirectURIs {
		if u, err := url.Parse(raw); err == nil {
			redirectURIs = append(redirectURIs, *u)
		}
	}
	return &genapi.OIDCClient{
		ID:                      c.ID,
		RedirectUris:            redirectURIs,
		GrantTypes:              c.GrantTypes,
		ResponseTypes:           c.ResponseTypes,
		Scope:                   c.ScopeIDs,
		TokenEndpointAuthMethod: c.TokenAuthnMethod,
		CreatedAt:               int64(c.CreatedAtTimestamp),
		ExpiresAt:               int64(c.ExpiresAtTimestamp),
	}
}

func urisToStrings(uris []url.URL) []string {
	out := make([]string, len(uris))
	for i, u := range uris {
		out[i] = u.String()
	}
	return out
}

func toUser(res *appusers.GetRes) *genapi.User {
	return &genapi.User{
		ID:       res.ID,
		Username: res.Username,
		Locale:   res.Locale,
	}
}

func updateOptsFromReq(req *genapi.UpdateUserRequest) appusers.UpdateOpts {
	return appusers.UpdateOpts{
		Username: optStringToPtr(req.Username),
		Locale:   optStringToPtr(req.Locale),
	}
}

func optStringToPtr(opt genapi.OptString) *string {
	if !opt.IsSet() {
		return nil
	}
	v := opt.Value
	return &v
}
