package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	api "github.com/4nd3r5on/oidc-serv/pkg/api"
)

func main() {
	log.SetFlags(0)

	fs := flag.NewFlagSet("oidc-adm", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: oidc-adm [-key KEY] [-url URL] <command> [args]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  clients  manage OIDC clients")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "flags:")
		fs.PrintDefaults()
	}

	key := fs.String("key", "", "admin API key (overrides OIDC_ADM_KEY env var)")
	serverURL := fs.String("url", "", "server base URL (overrides OIDC_ADM_URL; default: http://localhost:9090/api/v1)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	adminKey := *key
	if adminKey == "" {
		adminKey = os.Getenv("OIDC_ADM_KEY")
	}

	baseURL := *serverURL
	if baseURL == "" {
		baseURL = os.Getenv("OIDC_ADM_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:9090/api/v1"
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(0)
	}

	switch args[0] {
	case "clients":
		runClients(args[1:], adminKey, baseURL)
	default:
		log.Fatalf("unknown command: %s", args[0])
	}
}

// mustClient validates the admin key and returns a ready API client.
// Called only when a subcommand is about to make an actual API call.
func mustClient(adminKey, baseURL string) *api.Client {
	if adminKey == "" {
		log.Fatal("admin key required: set OIDC_ADM_KEY or pass -key")
	}
	c, err := newAdminClient(baseURL, adminKey)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	return c
}

func newAdminClient(serverURL, adminKey string) (*api.Client, error) {
	return api.NewClient(serverURL, &adminSecurity{key: adminKey})
}

// parseFlags parses a FlagSet and exits 0 on -help, 1 on other errors.
func parseFlags(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		log.Fatal(err)
	}
}
