package keymanager

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

type AlgoEdDSA struct{}

func (AlgoEdDSA) Name() string {
	return string(jose.EdDSA)
}

func (AlgoEdDSA) IsSymmetric() bool {
	return false
}

func (AlgoEdDSA) ParseRawConfig(rawConfig RawConfig) (*Config, error) {
	cfg := Config{Algorithm: string(jose.EdDSA)}

	if rawConfig.PublicKeyPEM != "" {
		pubKey, err := parseECDSAPublicKey(rawConfig.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		cfg.PublicKey = pubKey
	}

	if rawConfig.PrivateKeyPEM != "" {
		privKey, err := parseECDSAPrivateKey(rawConfig.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		cfg.PrivateKey = privKey
		if cfg.PublicKey == nil {
			cfg.PublicKey = privKey.Public()
		}
	}

	return &cfg, nil
}

func parseECDSAPublicKey(pemData string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKIX public key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not ECDSA")
	}

	return ecdsaKey, nil
}

func parseECDSAPrivateKey(pemData string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS8 first
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ecdsaKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not ECDSA")
		}
		return ecdsaKey, nil
	}

	// Try EC-specific format
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC private key: %w", err)
	}

	return key, nil
}
