package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"

	appprovider "github.com/4nd3r5on/oidc-serv/internal/app/provider"
	"github.com/4nd3r5on/oidc-serv/internal/keymanager"
)

func initProvider(
	app *App,
	repos *Repos,
	jwtConfigPath string,
	issuer string,
	logger *slog.Logger,
) (*appprovider.Provider, error) {
	jwtConfig, err := resolveJWTConfig(jwtConfigPath, "")
	if err != nil {
		return nil, err
	}

	jwk, err := buildJWK(jwtConfig, logger)
	if err != nil {
		return nil, err
	}

	jwks := goidc.JSONWebKeySet{Keys: []goidc.JSONWebKey{jwk}}
	op, _ := provider.New(
		goidc.ProfileOpenID,
		issuer,
		func(_ context.Context) (goidc.JSONWebKeySet, error) {
			return jwks, nil
		},
	)

	return appprovider.New(
		op,
		appprovider.StorageConfig{
			Clients:        repos.Clients,
			Sessions:       repos.Sessions,
			Grants:         repos.Grants,
			Tokens:         repos.Tokens,
			LogoutSessions: repos.LogoutSessions,
		},
		repos.Users,
		app.MatchUserPass.MatchUserPass,
	)
}

func buildJWK(cfg *keymanager.Config, logger *slog.Logger) (goidc.JSONWebKey, error) {
	jwk := goidc.JSONWebKey{
		Algorithm: cfg.Algorithm,
		KeyID:     "key_id", // TODO: make configurable
	}

	if keymanager.Algorithms[cfg.Algorithm].IsSymmetric() {
		if cfg.SecretKey == nil {
			return goidc.JSONWebKey{}, errors.New("secret key required for symmetric algorithm")
		}
		jwk.Key = cfg.SecretKey
		return jwk, nil
	}

	// asymmetric
	if cfg.SecretKey != nil {
		jwk.Key = cfg.SecretKey
	} else if cfg.PublicKey != nil {
		logger.Warn("private key not provided for asymmetric algorithm", "algorithm", cfg.Algorithm)
		jwk.Key = cfg.PublicKey
	} else {
		return goidc.JSONWebKey{}, errors.New("public or private key required for asymmetric algorithm")
	}
	return jwk, nil
}
