package keymanager

import (
	"github.com/go-jose/go-jose/v4"
)

// Config holds configuration for creating a TokenManager
type Config struct {
	// Algorithm specifies the signing algorithm to use
	Algorithm string

	// SecretKey is used for symmetric algorithms (HS256)
	// Should be at least 32 bytes for HS256
	SecretKey []byte

	// PrivateKey is used for asymmetric algorithms (ES256, EdDSA) for signing
	// Can be *ecdsa.PrivateKey for ES256 or ed25519.PrivateKey for EdDSA
	PrivateKey any

	// PublicKey is used for asymmetric algorithms (ES256, EdDSA) for verification
	// Can be *ecdsa.PublicKey for ES256 or ed25519.PublicKey for EdDSA
	// If PrivateKey is provided, PublicKey will be derived automatically
	PublicKey any
}

type AlgoVariant interface {
	ParseRawConfig(RawConfig) (*Config, error)
	IsSymmetric() bool
	Name() string
}

var Algorithms = map[string]AlgoVariant{
	string(jose.HS256): AlgoHS256{},
	string(jose.RS256): AlgoRS256{},
	string(jose.EdDSA): AlgoEdDSA{},
	string(jose.ES256): AlgoES256{},
}

func GetSupportedAlgorithms() []string {
	dst := make([]string, 0, len(Algorithms))
	for key := range Algorithms {
		dst = append(dst, key)
	}
	return dst
}
