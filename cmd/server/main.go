package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	appprovider "github.com/4nd3r5on/oidc-serv/internal/app/provider"
	"github.com/4nd3r5on/oidc-serv/internal/config"
	"github.com/4nd3r5on/oidc-serv/internal/keymanager"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/rs/cors"
)

func main() {
	_ = context.Background()
	serverAddr := getEnv(config.EnvServerAddr, ":9090")
	issuer := ""
	env := config.GetEnvironment()
	// TODO: Use config path
	jwtConfigPath := flag.String("jwt-cfg", "./jwt_config.yml", "JWT Config file path")
	flag.Parse()

	if env == config.EnvironmentUnknown {
		log.Fatalf(
			"missconfigured or missing required enviromnment variable %s",
			config.EnvEnvironment,
		)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: env.LogLevel(),
	}))
	slog.SetDefault(logger)

	jwtConfig, err := resolveJWTConfig(*jwtConfigPath, "")
	if err != nil {
		log.Fatalf("failed to resolve JWT config: %v", err)
	}
	jwk := goidc.JSONWebKey{
		Algorithm: jwtConfig.Algorithm,
		KeyID:     "key_id", // TODO: Change
	}
	jwtIsSymmetricAlgo := keymanager.Algorithms[jwtConfig.Algorithm].IsSymmetric()
	if jwtIsSymmetricAlgo {
		if jwtConfig.SecretKey == nil {
			// TODO: error
			panic("err")
		}
		jwk.Key = jwtConfig.SecretKey
	} else {
		if jwtConfig.SecretKey != nil {
			jwk.Key = jwtConfig.SecretKey
		} else if jwtConfig.PublicKey == nil {
			// TODO: error
			panic("err")
		}
		logger.Warn(
			"private key not provided for asymmetric algorithm",
			"algorithm", jwtConfig.Algorithm,
		)
		jwk.Key = jwtConfig.PublicKey
	}

	jwks := goidc.JSONWebKeySet{Keys: []goidc.JSONWebKey{jwk}}
	op, _ := provider.New(
		goidc.ProfileOpenID,
		issuer,
		func(_ context.Context) (goidc.JSONWebKeySet, error) {
			return jwks, nil
		},
	)

	// TODO: Init users APP Layer
	// TODO: Init storage

	appProvider, err := appprovider.New(op, appprovider.StorageConfig{}, nil)
	if err != nil {
		log.Fatalf(
			"failed to init app layer provider wrapper %v", err,
		)
	}

	mux := http.NewServeMux()

	// TODO: Init API Hanlders
	// TODO: Init Security Handlers

	apiPrefix := "/api/v1/"
	mux.Handle(apiPrefix, http.StripPrefix(apiPrefix, nil))
	mux.Handle("/", appProvider.Handler())
	httpHandler := cors.AllowAll().Handler(mux)

	server := &http.Server{
		Addr:        serverAddr,
		Handler:     httpHandler,
		ReadTimeout: 5 * time.Second,
	}
	logger.Info("Running server", "addr", serverAddr)
	server.ListenAndServe()
}
