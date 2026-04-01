openapi-gen:
    npx @redocly/cli bundle ./pkg/api/openapi/api.yml --ext yml -o ./pkg/api/api.yml
    go generate ./...

run-dev:
    reflex -r \\.go$ -s -- sh -c go run ./cmd/server
