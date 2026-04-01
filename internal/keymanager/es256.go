package keymanager

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

type AlgoES256 struct{}

func (AlgoES256) Name() string {
	return string(jose.ES256)
}

func (AlgoES256) IsSymmetric() bool {
	return false
}

func (AlgoES256) ParseRawConfig(rawConfig RawConfig) (*Config, error) {
	var err error
	cfg := Config{Algorithm: string(jose.ES256)}

	if rawConfig.PublicKeyPEM != "" {
		cfg.PublicKey, err = parseEd25519PublicKey(rawConfig.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
	}

	if rawConfig.PrivateKeyPEM != "" {
		cfg.PrivateKey, err = parseEd25519PrivateKey(rawConfig.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
	}

	return &cfg, nil
}

func parseEd25519PublicKey(pemData string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKIX public key: %w", err)
	}

	ed25519Key, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not Ed25519")
	}

	return ed25519Key, nil
}

func parseEd25519PrivateKey(pemData string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}

	ed25519Key, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("PKCS8 key is not Ed25519")
	}

	return ed25519Key, nil
}
