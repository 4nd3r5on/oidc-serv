package main

import (
	"context"
	"encoding/hex"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/4nd3r5on/oidc-serv/internal/config"
	inframemory "github.com/4nd3r5on/oidc-serv/internal/infra/memory"
	"github.com/4nd3r5on/oidc-serv/internal/infra/postgres"
	infraredis "github.com/4nd3r5on/oidc-serv/internal/infra/redis"
	"github.com/4nd3r5on/oidc-serv/pkg/db"
)

// Repos holds all repository implementations.
type Repos struct {
	Users          *postgres.UserRepo
	Clients        *inframemory.ClientRepoCached
	Sessions       *infraredis.SessionRepo
	Grants         *infraredis.GrantRepoCached
	Tokens         *infraredis.TokenRepo
	LogoutSessions *infraredis.LogoutSessionRepo
	UserSessions   *infraredis.UserSessionRepo
}

func mustLoadEncKey() []byte {
	encKeyHex := getEnv(config.EnvEncryptionKey, "")
	encKey, err := hex.DecodeString(encKeyHex)
	if err != nil || len(encKey) != 32 {
		log.Fatalf("%s must be a 64-char hex string (32 bytes)", config.EnvEncryptionKey)
	}
	return encKey
}

func mustConnectDB(ctx context.Context) *pgxpool.Pool {
	dbURL := getEnv(config.EnvDatabaseURL, "")
	if dbURL == "" {
		log.Fatalf("missing required env var %s", config.EnvDatabaseURL)
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	return pool
}

func mustLoadAdminAPIKey() string {
	key := getEnv(config.EnvAdminAPIKey, "")
	if key == "" {
		log.Fatalf("missing required env var %s", config.EnvAdminAPIKey)
	}
	return key
}

func mustConnectRedis() *redis.Client {
	redisURL := getEnv(config.EnvRedisURL, "")
	if redisURL == "" {
		log.Fatalf("missing required env var %s", config.EnvRedisURL)
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("parse redis URL: %v", err)
	}
	return redis.NewClient(opts)
}

func initRepos(q *db.Queries, redisClient *redis.Client, encKey []byte) *Repos {
	clientRepo := inframemory.NewClientRepoCached(postgres.NewClientRepo(q, encKey))

	return &Repos{
		Users:          postgres.NewUserRepo(q),
		Clients:        clientRepo,
		Sessions:       infraredis.NewSessionRepo(redisClient),
		Grants:         infraredis.NewGrantRepoCached(postgres.NewGrantRepo(q), redisClient),
		Tokens:         infraredis.NewTokenRepo(redisClient),
		LogoutSessions: infraredis.NewLogoutSessionRepo(redisClient),
		UserSessions:   infraredis.NewUserSessionRepo(redisClient),
	}
}
