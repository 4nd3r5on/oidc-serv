# === BUILDING BACKEND ===
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY pkg pkg

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -ldflags="-s -w" -o ./bin/server ./cmd/server

# === RUNTIME ===
FROM gcr.io/distroless/static-debian13:nonroot

ENV ENVIRONMENT=PROD

WORKDIR /app
COPY --from=builder /app/bin/server ./bin/server

EXPOSE 9090
ENTRYPOINT ["./bin/server"]
