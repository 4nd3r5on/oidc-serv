package clients

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/luikyv/go-oidc/pkg/goidc"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

var validGrantTypes = []string{
	string(goidc.GrantAuthorizationCode),
	string(goidc.GrantRefreshToken),
	string(goidc.GrantClientCredentials),
	string(goidc.GrantImplicit),
	string(goidc.GrantJWTBearer),
}

var validResponseTypes = []string{
	string(goidc.ResponseTypeCode),
	string(goidc.ResponseTypeIDToken),
	string(goidc.ResponseTypeToken),
	string(goidc.ResponseTypeCodeAndIDToken),
	string(goidc.ResponseTypeCodeAndToken),
	string(goidc.ResponseTypeIDTokenAndToken),
	string(goidc.ResponseTypeCodeAndIDTokenAndToken),
}

var validAuthnMethods = []string{
	string(goidc.AuthnMethodNone),
	string(goidc.AuthnMethodSecretBasic),
	string(goidc.AuthnMethodSecretPost),
	string(goidc.AuthnMethodSecretJWT),
	string(goidc.AuthnMethodPrivateKeyJWT),
	string(goidc.AuthnMethodTLS),
	string(goidc.AuthnMethodSelfSignedTLS),
}

func validateID(id string) error {
	if id == "" {
		return errs.Mark(fmt.Errorf("client id must not be empty"), errs.ErrInvalidArgument)
	}
	return nil
}

func validateRedirectURIs(uris []string) error {
	if len(uris) == 0 {
		return errs.Mark(fmt.Errorf("at least one redirect URI is required"), errs.ErrInvalidArgument)
	}
	for _, raw := range uris {
		u, err := url.ParseRequestURI(raw)
		if err != nil || !u.IsAbs() {
			return errs.Mark(fmt.Errorf("invalid redirect URI %q", raw), errs.ErrInvalidArgument)
		}
	}
	return nil
}

func validateGrantTypes(types []string) error {
	for _, gt := range types {
		if !slices.Contains(validGrantTypes, gt) {
			return errs.Mark(fmt.Errorf("unsupported grant type %q", gt), errs.ErrInvalidArgument)
		}
	}
	return nil
}

func validateResponseTypes(types []string) error {
	for _, rt := range types {
		if !slices.Contains(validResponseTypes, rt) {
			return errs.Mark(fmt.Errorf("unsupported response type %q", rt), errs.ErrInvalidArgument)
		}
	}
	return nil
}

func validateTokenAuthnMethod(method string) error {
	if method == "" {
		return nil
	}
	if !slices.Contains(validAuthnMethods, method) {
		return errs.Mark(fmt.Errorf("unsupported token endpoint auth method %q", method), errs.ErrInvalidArgument)
	}
	return nil
}
