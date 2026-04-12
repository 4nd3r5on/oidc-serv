package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	api "github.com/4nd3r5on/oidc-serv/pkg/api"
)

// stringSlice accumulates repeated flag values.
type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

// ---- clientsCmd -------------------------------------------------------

type clientsCmd struct {
	subcmds map[string]Cmd
}

func newClientsCmd() *clientsCmd {
	return &clientsCmd{subcmds: map[string]Cmd{
		"create": &clientsCreateCmd{},
		"get":    &clientsGetCmd{},
		"delete": &clientsDeleteCmd{},
	}}
}

func (c *clientsCmd) Short() string { return "manage OIDC clients" }

func (c *clientsCmd) Help(w io.Writer, prefix string) {
	fmt.Fprintf(w, "usage: %s clients <subcommand>\n\n", prefix)
	fmt.Fprintln(w, "subcommands:")
	for _, name := range sortedKeys(c.subcmds) {
		fmt.Fprintf(w, "  %-8s  %s\n", name, c.subcmds[name].Short())
	}
}

func (c *clientsCmd) Exec(ctx context.Context, w io.Writer, args []string) error {
	if len(args) == 0 {
		c.Help(w, "oidc-adm")
		return nil
	}
	sub, ok := c.subcmds[args[0]]
	if !ok {
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
	return sub.Exec(ctx, w, args[1:])
}

// ---- clientsCreateCmd -------------------------------------------------

type clientsCreateCmd struct{}

func (c *clientsCreateCmd) Short() string { return "create a new OIDC client" }

func (c *clientsCreateCmd) Help(w io.Writer, prefix string) {
	fmt.Fprintf(w, "usage: %s clients create -id ID -redirect-uri URI [flags]\n", prefix)
}

func (c *clientsCreateCmd) Exec(ctx context.Context, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("clients create", flag.ContinueOnError)
	fs.SetOutput(w)
	id := fs.String("id", "", "client ID (required)")
	secret := fs.String("secret", "", "plaintext secret (randomly generated if omitted)")
	scope := fs.String("scope", "", `space-separated scopes, e.g. "openid profile"`)
	authMethod := fs.String("auth-method", "", "token_endpoint_auth_method (default: client_secret_basic)")
	var redirectURIs, grantTypes, responseTypes stringSlice
	fs.Var(&redirectURIs, "redirect-uri", "redirect URI (repeatable, required)")
	fs.Var(&grantTypes, "grant-type", "grant type (repeatable; default: authorization_code)")
	fs.Var(&responseTypes, "response-type", "response type (repeatable; default: code)")
	fs.Usage = func() { c.Help(w, "oidc-adm"); fmt.Fprintln(w, "\nflags:"); fs.PrintDefaults() }

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *id == "" {
		return fmt.Errorf("-id is required")
	}
	if len(redirectURIs) == 0 {
		return fmt.Errorf("-redirect-uri is required")
	}

	parsedURIs := make([]url.URL, 0, len(redirectURIs))
	for _, raw := range redirectURIs {
		u, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("invalid redirect URI %q: %w", raw, err)
		}
		parsedURIs = append(parsedURIs, *u)
	}

	req := &api.CreateClientRequest{ID: *id, RedirectUris: parsedURIs}
	if len(grantTypes) > 0 {
		req.GrantTypes = []string(grantTypes)
	}
	if len(responseTypes) > 0 {
		req.ResponseTypes = []string(responseTypes)
	}
	if *secret != "" {
		req.Secret = api.NewOptString(*secret)
	}
	if *scope != "" {
		req.Scope = api.NewOptString(*scope)
	}
	if *authMethod != "" {
		req.TokenEndpointAuthMethod = api.NewOptString(*authMethod)
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	res, err := client.CreateClient(ctx, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	switch res := res.(type) {
	case *api.CreateClientResponse:
		fmt.Fprintf(w, "client created\n  id:     %s\n  secret: %s\n  (store the secret securely — returned only once)\n", res.ID, res.Secret)
	case *api.CreateClientBadRequest:
		return fmt.Errorf("bad request: %s", res.Error)
	case *api.CreateClientUnauthorized:
		return fmt.Errorf("unauthorized: %s", res.Error)
	default:
		return fmt.Errorf("unexpected response type: %T", res)
	}
	return nil
}

// ---- clientsGetCmd ----------------------------------------------------

type clientsGetCmd struct{}

func (c *clientsGetCmd) Short() string { return "get an OIDC client by ID" }

func (c *clientsGetCmd) Help(w io.Writer, prefix string) {
	fmt.Fprintf(w, "usage: %s clients get <clientId>\n", prefix)
}

func (c *clientsGetCmd) Exec(ctx context.Context, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("clients get", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.Usage = func() { c.Help(w, "oidc-adm") }

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() < 1 {
		c.Help(w, "oidc-adm")
		return nil
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	res, err := client.GetClientById(ctx, api.GetClientByIdParams{ClientId: fs.Arg(0)})
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	switch res := res.(type) {
	case *api.OIDCClient:
		uriStrs := make([]string, 0, len(res.RedirectUris))
		for _, u := range res.RedirectUris {
			uriStrs = append(uriStrs, u.String())
		}
		fmt.Fprintf(w, "id:                         %s\n", res.ID)
		fmt.Fprintf(w, "scope:                      %s\n", res.Scope)
		fmt.Fprintf(w, "grant_types:                %s\n", strings.Join(res.GrantTypes, ", "))
		fmt.Fprintf(w, "response_types:             %s\n", strings.Join(res.ResponseTypes, ", "))
		fmt.Fprintf(w, "token_endpoint_auth_method: %s\n", res.TokenEndpointAuthMethod)
		fmt.Fprintf(w, "redirect_uris:              %s\n", strings.Join(uriStrs, ", "))
		fmt.Fprintf(w, "created_at:                 %s\n", time.Unix(res.CreatedAt, 0).UTC().Format(time.RFC3339))
		if res.ExpiresAt != 0 {
			fmt.Fprintf(w, "expires_at:                 %s\n", time.Unix(res.ExpiresAt, 0).UTC().Format(time.RFC3339))
		}
	case *api.GetClientByIdNotFound:
		return fmt.Errorf("not found: %s", res.Error)
	case *api.GetClientByIdUnauthorized:
		return fmt.Errorf("unauthorized: %s", res.Error)
	default:
		return fmt.Errorf("unexpected response type: %T", res)
	}
	return nil
}

// ---- clientsDeleteCmd -------------------------------------------------

type clientsDeleteCmd struct{}

func (c *clientsDeleteCmd) Short() string { return "delete an OIDC client" }

func (c *clientsDeleteCmd) Help(w io.Writer, prefix string) {
	fmt.Fprintf(w, "usage: %s clients delete <clientId>\n", prefix)
}

func (c *clientsDeleteCmd) Exec(ctx context.Context, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("clients delete", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.Usage = func() { c.Help(w, "oidc-adm") }

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() < 1 {
		c.Help(w, "oidc-adm")
		return nil
	}

	id := fs.Arg(0)
	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	res, err := client.DeleteClient(ctx, api.DeleteClientParams{ClientId: id})
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	switch res := res.(type) {
	case *api.DeleteClientNoContent:
		fmt.Fprintf(w, "client %q deleted\n", id)
	case *api.DeleteClientNotFound:
		return fmt.Errorf("not found: %s", res.Error)
	case *api.DeleteClientUnauthorized:
		return fmt.Errorf("unauthorized: %s", res.Error)
	default:
		return fmt.Errorf("unexpected response type: %T", res)
	}
	return nil
}
