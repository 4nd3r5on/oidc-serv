# OIDC Server

My implementation of OIDC server.

## Why? 

It's about [KISS](https://en.wikipedia.org/wiki/KISS_principle) and freedom

I just implement the things I need, without bloat and overhead that big solutions like Keycloak or Authentik create.

I can implement whatever I need:

- Create and change scopes and claims however I need
- Create API methods and webhooks I need
- Make integrations with anything I need
- I create the architecture, so project might be tweaked to my resources

The only limit is my skill, not some other maintener

## Dependencies

- OpenSSL -- generating JWT keys

### Development 

- Migrate
- SQLc
- GolangCI Lint

## Config

Supported JWT algorithms: `EdDSA`, `ES256`, `HS256`, `RS256`

### JWT

#### Generate keys 
```sh
# TODO: OpenSSL commands here
```

#### Create config
You can go to your config with `nvim ./jwt_config.yml` 

```yml
algorithm: "EdDSA"
private_key_path: ./jwt.pem
public_key_path: ./jwt_pub.pem
```

```yml
algorithm: "HS256"
secret_key: ">> std base64 encoded key here <<"
```

## Scopes

To be extended

#### default

By default we include `sub` and `nonce`

#### `core`

Includes `username` and `locale`

