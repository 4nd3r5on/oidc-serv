package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"

	internalapi "github.com/4nd3r5on/oidc-serv/internal/api"
	"github.com/4nd3r5on/oidc-serv/internal/config"
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
	"github.com/4nd3r5on/oidc-serv/pkg/db"
)

func main() {
	ctx := context.Background()

	jwtConfigPath := flag.String("jwt-cfg", "./jwt_config.yml", "JWT Config file path")
	flag.Parse()

	env := config.GetEnvironment()
	if env == config.EnvironmentUnknown {
		log.Fatalf("misconfigured or missing required environment variable %s", config.EnvEnvironment)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: env.LogLevel(),
	}))
	slog.SetDefault(logger)

	encKey := mustLoadEncKey()

	pool := mustConnectDB(ctx)
	defer pool.Close()

	redisClient := mustConnectRedis()
	defer redisClient.Close()

	repos := initRepos(ctx, db.New(pool), redisClient, encKey, logger)
	app := initApp(repos, logger)

	appProvider, err := initProvider(app, repos, *jwtConfigPath, "", logger)
	if err != nil {
		log.Fatalf("init OIDC provider: %v", err)
	}

	securityHandler := &internalapi.SecurityHandler{
		TMB:      app.TMBVerifier,
		Session:  app.SessionVerifier,
		AdminKey: mustLoadAdminAPIKey(),
	}
	handlers := &internalapi.Handlers{
		ClientCreate: app.CreateClient,
		ClientGet:    app.GetClient,
		ClientDelete: app.DeleteClient,

		Create:        app.CreateUser,
		GetByID:       app.GetUser,
		GetByUsername: app.GetUserByUsername,
		Me:            app.Me,
	}
	apiServer, err := genapi.NewServer(handlers, securityHandler)
	if err != nil {
		log.Fatalf("init API server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiServer))
	mux.Handle("/", appProvider.Handler())

	server := &http.Server{
		Addr: getEnv(config.EnvServerAddr, ":9090"),
		// TODO: Add proper CORS configuration
		Handler:     cors.AllowAll().Handler(mux),
		ReadTimeout: 5 * time.Second,
	}
	logger.Info("running server", "addr", server.Addr)
	log.Fatal(server.ListenAndServe())
}
