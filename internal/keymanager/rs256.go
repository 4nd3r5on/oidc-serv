package keymanager

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

type AlgoRS256 struct{}

func (AlgoRS256) Name() string {
	return string(jose.RS256)
}

func (AlgoRS256) IsSymmetric() bool {
	return false
}

func (AlgoRS256) ParseRawConfig(rawConfig RawConfig) (*Config, error) {
	cfg := Config{Algorithm: string(jose.RS256)}

	if rawConfig.PublicKeyPEM != "" {
		pubKey, err := parseRSAPublicKey(rawConfig.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		cfg.PublicKey = pubKey
	}

	if rawConfig.PrivateKeyPEM != "" {
		privKey, err := parseRSAPrivateKey(rawConfig.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		cfg.PrivateKey = privKey
		if cfg.PublicKey == nil {
			cfg.PublicKey = &privKey.PublicKey
		}
	}

	return &cfg, nil
}

func parseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKIX public key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return rsaKey, nil
}

func parseRSAPrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS8 first
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
		return rsaKey, nil
	}

	// Try PKCS1 format
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS1 private key: %w", err)
	}

	return key, nil
}
