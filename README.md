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
|----------|----------|-------------|
| `ENVIRONMENT` | yes | `prod`, `dev`, or `test` |
| `DATABASE_URL` | yes | PostgreSQL connection URL |
| `REDIS_URL` | yes | Redis connection URL, e.g. `redis://localhost:6379` |
| `ENCRYPTION_KEY` | yes | 64-char hex string (32 bytes) used for AES-256-GCM encryption at rest — generate with `openssl rand -hex 32` |
| `ADMIN_API_KEY` | yes | Static key for client management endpoints — passed as `X-Admin-Key` header |
| `SERVER_ADDR` | no | Listen address, default `:9090` |

### Encryption key

```sh
# TODO: DB_URL, REDIS_URL
echo "ENCRYPTION_KEY=$(openssl rand -hex 32)" >> .env
echo "ADMIN_API_KEY=$(openssl rand -hex 32)" >> .env
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
algorithm: "HS256"
```
then execute to generate a secret key
```sh
echo "secret_key: $(openssl rand -base64 32)" >> jwt_config.yml
```

## Managing clients

Clients are managed via the `/api/v1/clients` endpoints, protected by `ADMIN_API_KEY`
passed as the `X-Admin-Key` header.

The secret is never stored in plaintext — it is encrypted with AES-256-GCM using `ENCRYPTION_KEY`
before being written, matching the server's [`Encrypt`](internal/infra/crypto/crypto.go) helper.

### Create

```sh
curl -s -X POST http://localhost:9090/api/v1/clients \
  -H "X-Admin-Key: $ADMIN_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-client",
    "redirect_uris": ["http://localhost:8080/callback"],
    "scope": "openid profile"
  }' | jq .
```

```json
{
  "id": "my-client",
  "secret": "<plaintext secret>"
}
```

Save the `secret` — it is returned only once and cannot be recovered from the database.

When a field is omitted from the request body, the server fills in these defaults:

| Field | Default |
|---|---|
| `grant_types` | `["authorization_code"]` |
| `response_types` | `["code"]` |
| `token_endpoint_auth_method` | `"client_secret_basic"` |
| `secret` | random 64-char hex string |

#### Grant types

| Value | Description |
|---|---|
| `authorization_code` | Standard redirect-based flow. User is sent to the authorization endpoint, authenticates, and the client exchanges the returned code for tokens. Default and recommended for most clients. |
| `refresh_token` | Allows the client to obtain a new access token using a refresh token, without re-authenticating the user. Include alongside `authorization_code` when long-lived sessions are needed. |

### Get

```sh
curl -s http://localhost:9090/api/v1/clients/my-client \
  -H "X-Admin-Key: $ADMIN_API_KEY" | jq .
```

### Delete

```sh
curl -s -X DELETE http://localhost:9090/api/v1/clients/my-client \
  -H "X-Admin-Key: $ADMIN_API_KEY"
```

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
