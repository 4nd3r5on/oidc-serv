# OIDC Server

A minimal, self-hosted OpenID Connect provider built on top of [go-oidc](https://github.com/luikyv/go-oidc).

## Why?

It's about [KISS](https://en.wikipedia.org/wiki/KISS_principle) and freedom.

I implement only what I need, without introducing overhead of services like Keycloak or Authentik.  
That means full control over scopes, claims, API endpoints, integrations, and architecture.  
The only limit is my own skill.

## Requirements

**Runtime**
- PostgreSQL
- Redis

**Development**
- [migrate](https://github.com/golang-migrate/migrate) — database migrations
- [sqlc](https://sqlc.dev) — query code generation
- [golangci-lint](https://golangci-lint.run) — linting
- [reflex](https://github.com/cespare/reflex) — live reload (`just run-dev`)

## Configuration

| Variable | Required | Description |
|---|---|---|
| `ENVIRONMENT` | yes | `prod`, `dev`, or `test` |
| `DATABASE_URL` | yes | PostgreSQL connection URL |
| `REDIS_URL` | yes | Redis connection URL, e.g. `redis://localhost:6379` |
| `ENCRYPTION_KEY` | yes | 64-char hex string (32 bytes) used for AES-256-GCM encryption at rest — generate with `openssl rand -hex 32` |
| `SERVER_ADDR` | no | Listen address, default `:9090` |

### Encryption key

```sh
# TODO: DB_URL, REDIS_URL
echo "ENCRYPTION_KEY=$(openssl rand -hex 32)" >> .env
```

### JWT

The server is configured via a `jwt_config.yml` file (path overridable with `--jwt-cfg`).

**Supported algorithms:** `EdDSA`, `ES256`, `RS256`, `HS256`

#### Asymmetric (EdDSA / ES256 / RS256)

```sh
# EdDSA
openssl genpkey -algorithm ed25519 -out jwt.pem
openssl pkey -in jwt.pem -pubout -out jwt_pub.pem
```

Then create a `jwt_config.yml` file containing
```yml
algorithm: "EdDSA"
private_key_path: ./jwt.pem
public_key_path: ./jwt_pub.pem
```

#### Symmetric (HS256)

Create `jwt_config.yml` containing
```yml
# jwt_config.yml
algorithm: "HS256"
secret_key: "<base64-encoded secret>"
```

If u want to generate the secret key and put it right into the config u can use
```sh
yq -i '.secret_key = "'$(openssl rand -base64 32)'"' jwt_config.yml
```

## Scopes

| Scope | Claims added |
|---|---|
| *(default)* | `sub`, `nonce` |
| `core` | `username`, `locale` |

## Development

```sh
# live reload
just run-dev

# regenerate sqlc queries after changing SQL files
sqlc generate

# regenerate API stubs after changing the OpenAPI spec
just openapi-gen
```
