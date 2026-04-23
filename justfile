build-server:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server ./cmd/server

build-adm:
    go build -o ./bin/oidc-adm ./cmd/adm

build: build-server build-adm

openapi-gen:
    npx @redocly/cli bundle ./pkg/api/openapi/api.yml --ext yml -o ./pkg/api/api.yml
    go generate ./...

run-dev:
    ENVIRONMENT=DEV reflex -r \\.go$ -s -- sh -c go run ./cmd/server
