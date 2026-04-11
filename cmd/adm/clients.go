package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	api "github.com/4nd3r5on/oidc-serv/pkg/api"
)

// stringSlice accumulates repeated flag values.
type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

func runClients(args []string, adminKey, baseURL string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: oidc-adm clients <subcommand> [args]")
		fmt.Fprintln(os.Stderr, "subcommands: create, get, delete")
		os.Exit(0)
	}

	switch args[0] {
	case "create":
		clientsCreate(args[1:], adminKey, baseURL)
	case "get":
		clientsGet(args[1:], adminKey, baseURL)
	case "delete":
		clientsDelete(args[1:], adminKey, baseURL)
	default:
		log.Fatalf("unknown subcommand: %s", args[0])
	}
}

func clientsCreate(args []string, adminKey, baseURL string) {
	fs := flag.NewFlagSet("clients create", flag.ContinueOnError)
	id := fs.String("id", "", "client ID (required)")
	secret := fs.String("secret", "", "plaintext secret (randomly generated if omitted)")
	scope := fs.String("scope", "", `space-separated scopes, e.g. "openid profile"`)
	authMethod := fs.String("auth-method", "", "token_endpoint_auth_method (default: client_secret_basic)")

	var redirectURIs stringSlice
	var grantTypes stringSlice
	var responseTypes stringSlice
	fs.Var(&redirectURIs, "redirect-uri", "redirect URI (repeatable, required)")
	fs.Var(&grantTypes, "grant-type", "grant type (repeatable; default: authorization_code)")
	fs.Var(&responseTypes, "response-type", "response type (repeatable; default: code)")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: oidc-adm clients create -id ID -redirect-uri URI [flags]")
		fs.PrintDefaults()
	}
	parseFlags(fs, args)

	if *id == "" {
		log.Fatal("-id is required")
	}
	if len(redirectURIs) == 0 {
		log.Fatal("-redirect-uri is required")
	}

	parsedURIs := make([]url.URL, 0, len(redirectURIs))
	for _, raw := range redirectURIs {
		u, err := url.Parse(raw)
		if err != nil {
			log.Fatalf("invalid redirect URI %q: %v", raw, err)
		}
		parsedURIs = append(parsedURIs, *u)
	}

	req := &api.CreateClientRequest{
		ID:           *id,
		RedirectUris: parsedURIs,
	}
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

	client := mustClient(adminKey, baseURL)
	res, err := client.CreateClient(context.Background(), req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	switch res := res.(type) {
	case *api.CreateClientResponse:
		fmt.Printf("client created\n")
		fmt.Printf("  id:     %s\n", res.ID)
		fmt.Printf("  secret: %s\n", res.Secret)
		fmt.Printf("  (store the secret securely — returned only once)\n")
	case *api.CreateClientBadRequest:
		log.Fatalf("bad request: %s", res.Error)
	case *api.CreateClientUnauthorized:
		log.Fatalf("unauthorized: %s", res.Error)
	default:
		log.Fatalf("unexpected response type: %T", res)
	}
}

func clientsGet(args []string, adminKey, baseURL string) {
	fs := flag.NewFlagSet("clients get", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: oidc-adm clients get <clientId>")
	}
	parseFlags(fs, args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(0)
	}

	client := mustClient(adminKey, baseURL)
	res, err := client.GetClientById(context.Background(), api.GetClientByIdParams{
		ClientId: fs.Arg(0),
	})
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	switch res := res.(type) {
	case *api.OIDCClient:
		uriStrs := make([]string, 0, len(res.RedirectUris))
		for _, u := range res.RedirectUris {
			uriStrs = append(uriStrs, u.String())
		}
		fmt.Printf("id:                         %s\n", res.ID)
		fmt.Printf("scope:                      %s\n", res.Scope)
		fmt.Printf("grant_types:                %s\n", strings.Join(res.GrantTypes, ", "))
		fmt.Printf("response_types:             %s\n", strings.Join(res.ResponseTypes, ", "))
		fmt.Printf("token_endpoint_auth_method: %s\n", res.TokenEndpointAuthMethod)
		fmt.Printf("redirect_uris:              %s\n", strings.Join(uriStrs, ", "))
		fmt.Printf("created_at:                 %s\n", time.Unix(res.CreatedAt, 0).UTC().Format(time.RFC3339))
		if res.ExpiresAt != 0 {
			fmt.Printf("expires_at:                 %s\n", time.Unix(res.ExpiresAt, 0).UTC().Format(time.RFC3339))
		}
	case *api.GetClientByIdNotFound:
		log.Fatalf("not found: %s", res.Error)
	case *api.GetClientByIdUnauthorized:
		log.Fatalf("unauthorized: %s", res.Error)
	default:
		log.Fatalf("unexpected response type: %T", res)
	}
}

func clientsDelete(args []string, adminKey, baseURL string) {
	fs := flag.NewFlagSet("clients delete", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: oidc-adm clients delete <clientId>")
	}
	parseFlags(fs, args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(0)
	}

	id := fs.Arg(0)
	client := mustClient(adminKey, baseURL)
	res, err := client.DeleteClient(context.Background(), api.DeleteClientParams{
		ClientId: id,
	})
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	switch res := res.(type) {
	case *api.DeleteClientNoContent:
		fmt.Printf("client %q deleted\n", id)
	case *api.DeleteClientNotFound:
		log.Fatalf("not found: %s", res.Error)
	case *api.DeleteClientUnauthorized:
		log.Fatalf("unauthorized: %s", res.Error)
	default:
		log.Fatalf("unexpected response type: %T", res)
	}
}
