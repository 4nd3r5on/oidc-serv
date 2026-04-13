package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	api "github.com/4nd3r5on/oidc-serv/pkg/api"
)

type Cmd interface {
	// Short returns a one-line description of the command.
	Short() string
	// Help writes usage text and subcommands.
	// Flag help is appended separately via flag.FlagSet.PrintDefaults() in Exec.
	Help(ctx context.Context, prefix string)
	Exec(ctx context.Context, args []string) error
}

// rootCmd is the top-level command; it parses global flags and dispatches.
type rootCmd struct {
	subcmds map[string]Cmd
}

func (r *rootCmd) Short() string { return "OIDC server admin tool" }

func (r *rootCmd) Help(ctx context.Context, prefix string) {
	fmt.Printf("%susage: %s [-key KEY] [-url URL] <command> [args]\n\n", prefix, cmdLine(ctx))
	fmt.Println("commands:")
	for _, name := range sortedKeys(r.subcmds) {
		fmt.Printf("  %-8s  %s\n", name, r.subcmds[name].Short())
	}
}

func (r *rootCmd) Exec(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("oidc-adm", flag.ContinueOnError)
	key := fs.String("key", "", "admin API key (overrides OIDC_ADM_KEY env var)")
	serverURL := fs.String("url", "", "server base URL (overrides OIDC_ADM_URL; default: http://localhost:9090/api/v1)")
	fs.Usage = func() { r.Help(ctx, ""); fmt.Println("\nflags:"); fs.PrintDefaults() }

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
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
	ctx = ctxWithConfig(ctx, adminKey, baseURL)

	remaining := fs.Args()
	if len(remaining) == 0 {
		r.Help(ctx, "")
		return nil
	}

	sub, ok := r.subcmds[remaining[0]]
	if !ok {
		return fmt.Errorf("unknown command: %s", remaining[0])
	}
	return sub.Exec(ctxWithCmdName(ctx, remaining[0]), remaining[1:])
}

func main() {
	log.SetFlags(0)

	root := &rootCmd{
		subcmds: map[string]Cmd{
			"clients": newClientsCmd(),
		},
	}

	if err := root.Exec(ctxWithCmdName(context.Background(), "oidc-adm"), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

// apiClient builds an API client from config stored in ctx.
func apiClient(ctx context.Context) (*api.Client, error) {
	adminKey := adminKeyFromCtx(ctx)
	if adminKey == "" {
		return nil, fmt.Errorf("admin key required: set OIDC_ADM_KEY or pass -key")
	}
	return api.NewClient(baseURLFromCtx(ctx), &adminSecurity{key: adminKey})
}
