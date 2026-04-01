package keymanager

import (
	"encoding/base64"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

type AlgoHS256 struct{}

func (AlgoHS256) Name() string {
	return string(jose.HS256)
}

func (AlgoHS256) IsSymmetric() bool {
	return true
}

func (AlgoHS256) ParseRawConfig(rawConfig RawConfig) (*Config, error) {
	secretKey, err := base64.StdEncoding.DecodeString(rawConfig.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("decode secret key: %w", err)
	}

	if len(secretKey) < 32 {
		return nil, fmt.Errorf("secret key must be at least 32 bytes, got %d", len(secretKey))
	}

	return &Config{
		Algorithm: string(jose.HS256),
		SecretKey: secretKey,
	}, nil
}
