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

### Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ENVIRONMENT` | yes | `prod`, `dev`, or `test`. Controls log level: `prod` → INFO, `dev`/`test` → DEBUG |
| `DATABASE_URL` | yes | PostgreSQL connection URL, e.g. `postgres://user:pass@localhost:5432/oidc` |
| `REDIS_URL` | yes | Redis connection URL, e.g. `redis://localhost:6379` |
| `ENCRYPTION_KEY` | yes | 64-char hex string (32 bytes) for AES-256-GCM encryption at rest |
| `ADMIN_API_KEY` | yes | Static key for client management endpoints — passed as `X-Admin-Key` header |
| `SERVER_ADDR` | no | Listen address, e.g. `:8080` or `0.0.0.0:8080`. Default: `:9090` |

Minimal `.env` to get started:

```sh
ENVIRONMENT=dev
DATABASE_URL=postgres://user:pass@localhost:5432/oidc
REDIS_URL=redis://localhost:6379
ENCRYPTION_KEY=$(openssl rand -hex 32)
ADMIN_API_KEY=$(openssl rand -hex 32)
```

### Server flags

| Flag | Default | Description |
|------|---------|-------------|
| `--jwt-cfg` | `./jwt_config.yml` | Path to the JWT configuration file |

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
algorithm: "HS256"
```
then execute to generate a secret key
```sh
echo "secret_key: $(openssl rand -base64 32)" >> jwt_config.yml
```

## Managing clients

Clients are managed with `oidc-adm`, the admin CLI tool.

```sh
go build -o oidc-adm ./cmd/adm
```

Set the admin key (or pass it via `-key` on every call):

```sh
export OIDC_ADM_KEY="$ADMIN_API_KEY"
```

The server URL defaults to `http://localhost:9090/api/v1` and can be overridden with `-url` or `OIDC_ADM_URL`.

### Create

```sh
oidc-adm clients create -id my-client \
  -redirect-uri http://localhost:8080/callback \
  -scope "openid profile"
```

```
client created
  id:     my-client
  secret: <plaintext secret>
  (store the secret securely — returned only once)
```

Save the `secret` — it is returned only once and cannot be recovered from the database.

## Grant types

| Value | Description |
|---|---|
| `authorization_code` | Standard redirect-based flow. User is sent to the authorization endpoint, authenticates, and the client exchanges the returned code for tokens. Default and recommended for most clients. |
| `refresh_token` | Allows the client to obtain a new access token using a refresh token, without re-authenticating the user. Pass alongside `authorization_code` when long-lived sessions are needed. |

## Scopes

| Scope | Claims added |
|---|---|
| *(default)* | `sub`, `nonce` |
| `profile` | `preferred_username`, `locale` |

## Development

```sh
# live reload
just run-dev

# regenerate sqlc queries after changing SQL files
sqlc generate

# regenerate API stubs after changing the OpenAPI spec
just openapi-gen
```
