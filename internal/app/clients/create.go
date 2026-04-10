package clients

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

const secretBytesLen = 32

type CreateOpts struct {
	ID               string
	Secret           string // empty = generate
	RedirectURIs     []string
	GrantTypes       []string
	ResponseTypes    []string
	ScopeIDs         string
	TokenAuthnMethod string
}

type Create struct {
	Clients Repo
	Logger  *slog.Logger
}

func NewCreate(clients Repo, logger *slog.Logger) *Create {
	if logger == nil {
		logger = slog.Default()
	}
	return &Create{Clients: clients, Logger: logger}
}

func (c *Create) Create(ctx context.Context, opts CreateOpts) (CreateRes, error) {
	if err := validateID(opts.ID); err != nil {
		return CreateRes{}, err
	}
	if err := validateRedirectURIs(opts.RedirectURIs); err != nil {
		return CreateRes{}, err
	}
	if err := validateGrantTypes(opts.GrantTypes); err != nil {
		return CreateRes{}, err
	}
	if err := validateResponseTypes(opts.ResponseTypes); err != nil {
		return CreateRes{}, err
	}
	if err := validateTokenAuthnMethod(opts.TokenAuthnMethod); err != nil {
		return CreateRes{}, err
	}

	// The underlying repo uses upsert, so we must detect conflicts ourselves.
	if _, err := c.Clients.Client(ctx, opts.ID); err == nil {
		return CreateRes{}, errs.Mark(fmt.Errorf("client %q already exists", opts.ID), errs.ErrExists)
	} else if !errors.Is(err, errs.ErrNotFound) {
		c.Logger.ErrorContext(ctx, "create client: check existence error", "id", opts.ID, "error", err)
		return CreateRes{}, fmt.Errorf("check existence: %w", err)
	}

	secret := opts.Secret
	if secret == "" {
		var err error
		secret, err = generateSecret()
		if err != nil {
			c.Logger.ErrorContext(ctx, "create client: generate secret error", "error", err)
			return CreateRes{}, fmt.Errorf("generate secret: %w", err)
		}
	}

	grantTypes := opts.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{string(goidc.GrantAuthorizationCode)}
	}
	responseTypes := opts.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{string(goidc.ResponseTypeCode)}
	}
	tokenAuthnMethod := opts.TokenAuthnMethod
	if tokenAuthnMethod == "" {
		tokenAuthnMethod = string(goidc.AuthnMethodSecretBasic)
	}

	client := buildGoidcClient(opts.ID, secret, opts.RedirectURIs, grantTypes, responseTypes, opts.ScopeIDs, tokenAuthnMethod)
	if err := c.Clients.Save(ctx, client); err != nil {
		c.Logger.ErrorContext(ctx, "create client: repository error", "id", opts.ID, "error", err)
		return CreateRes{}, fmt.Errorf("repository: %w", err)
	}
	return CreateRes{ID: opts.ID, Secret: secret}, nil
}

func buildGoidcClient(id, secret string, redirectURIs, grantTypes, responseTypes []string, scopeIDs, tokenAuthnMethod string) *goidc.Client {
	gt := make([]goidc.GrantType, len(grantTypes))
	for i, s := range grantTypes {
		gt[i] = goidc.GrantType(s)
	}
	rt := make([]goidc.ResponseType, len(responseTypes))
	for i, s := range responseTypes {
		rt[i] = goidc.ResponseType(s)
	}
	return &goidc.Client{
		ID:                 id,
		Secret:             secret,
		CreatedAtTimestamp: int(time.Now().Unix()),
		ClientMeta: goidc.ClientMeta{
			RedirectURIs:     redirectURIs,
			GrantTypes:       gt,
			ResponseTypes:    rt,
			ScopeIDs:         scopeIDs,
			TokenAuthnMethod: goidc.AuthnMethod(tokenAuthnMethod),
		},
	}
}

func generateSecret() (string, error) {
	b := make([]byte, secretBytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
