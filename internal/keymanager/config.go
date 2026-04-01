// Package keymanager provides helpers for managing keys
package keymanager

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// FileConfig represents configuration that can be serialized to/from files
type FileConfig struct {
	Algorithm      string `json:"algorithm" yaml:"algorithm"`
	SecretKey      string `json:"secret_key,omitempty" yaml:"secret_key,omitempty"`             // base64 std encoding
	PrivateKeyPath string `json:"private_key_path,omitempty" yaml:"private_key_path,omitempty"` // path to PEM file
	PrivateKeyPEM  string `json:"private_key_pem,omitempty" yaml:"private_key_pem,omitempty"`   // PEM encoded
	PublicKeyPath  string `json:"public_key_path,omitempty" yaml:"public_key_path,omitempty"`   // path to PEM file
	PublicKeyPEM   string `json:"public_key_pem,omitempty" yaml:"public_key_pem,omitempty"`     // PEM encoded
}

// RawConfig holds string-encoded configuration before parsing
type RawConfig struct {
	Algorithm     string
	SecretKey     string // base64 std encoding
	PrivateKeyPEM string // PEM encoded
	PublicKeyPEM  string // PEM encoded
}

func requiredFieldsSymmetric() map[string]struct{} {
	return map[string]struct{}{"SecretKey": {}}
}

func requiredFieldsAsymmetric() map[string]struct{} {
	return map[string]struct{}{"PublicKeyPEM": {}}
}

func ResolveConfig(fileConfig FileConfig, envPrefix string, logger *slog.Logger) (*Config, error) {
	rawConfig, err := ResolveRawConfig(fileConfig, envPrefix, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to build raw config: %w", err)
	}
	config, err := RawConfigToConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert raw config to config: %w", err)
	}
	return config, nil
}

func ResolveRawConfig(fileConfig FileConfig, envPrefix string, logger *slog.Logger) (RawConfig, error) {
	rawConfig, err := FileConfigToRaw(fileConfig)
	if err != nil {
		return RawConfig{}, fmt.Errorf("failed to convert file config to raw config: %w", err)
	}
	rawConfig, err = ApplyEnv(rawConfig, envPrefix, logger)
	if err != nil {
		return RawConfig{}, fmt.Errorf("failed to apply env to raw config: %w", err)
	}
	return rawConfig, nil
}

// FileConfigToRaw converts FileConfig to RawConfig, reading PEM files if paths are provided
func FileConfigToRaw(fileConfig FileConfig) (RawConfig, error) {
	rawConfig := RawConfig{
		Algorithm:     fileConfig.Algorithm,
		SecretKey:     fileConfig.SecretKey,
		PrivateKeyPEM: fileConfig.PrivateKeyPEM,
		PublicKeyPEM:  fileConfig.PublicKeyPEM,
	}

	if fileConfig.PrivateKeyPath != "" {
		data, err := os.ReadFile(fileConfig.PrivateKeyPath)
		if err != nil {
			return RawConfig{}, fmt.Errorf("read private key file: %w", err)
		}
		rawConfig.PrivateKeyPEM = string(data)
	}

	if fileConfig.PublicKeyPath != "" {
		data, err := os.ReadFile(fileConfig.PublicKeyPath)
		if err != nil {
			return RawConfig{}, fmt.Errorf("read public key file: %w", err)
		}
		rawConfig.PublicKeyPEM = string(data)
	}

	return rawConfig, nil
}

// ApplyEnv overlays environment variables onto RawConfig
// prefix should end with "_" or be empty (e.g., "JWT_" or "")
// Returns warnings for conflicts where both file and env are set
func ApplyEnv(rawConfig RawConfig, envPrefix string, logger *slog.Logger) (RawConfig, error) {
	if envPrefix != "" && envPrefix[len(envPrefix)-1] != '_' {
		envPrefix += "_"
	}

	result := rawConfig

	envMap := map[string]struct {
		envKey string
		target *string
		source string
	}{
		"algorithm":   {envPrefix + "ALGORITHM", (*string)(&result.Algorithm), string(rawConfig.Algorithm)},
		"secret_key":  {envPrefix + "SECRET_KEY", &result.SecretKey, rawConfig.SecretKey},
		"private_key": {envPrefix + "PRIVATE_KEY", &result.PrivateKeyPEM, rawConfig.PrivateKeyPEM},
		"public_key":  {envPrefix + "PUBLIC_KEY", &result.PublicKeyPEM, rawConfig.PublicKeyPEM},
	}

	for name, mapping := range envMap {
		if val := os.Getenv(mapping.envKey); val != "" {
			if mapping.source != "" {
				logger.Warn(fmt.Sprintf(
					"env %s overrides file config for %s",
					mapping.envKey, name,
				))
			}
			*mapping.target = val
		}
	}

	return result, nil
}

// ValidateRawConfig ensures all required fields are present for the chosen algorithm
func ValidateRawConfig(rawConfig RawConfig) error {
	algo, exists := Algorithms[rawConfig.Algorithm]
	if !exists {
		return fmt.Errorf(
			"unsupported algorithm: %s. Supported algorithms: %s",
			rawConfig.Algorithm,
			strings.Join(GetSupportedAlgorithms(), ", "),
		)
	}

	var required map[string]struct{}
	isSymmetric := algo.IsSymmetric()
	if isSymmetric {
		required = requiredFieldsSymmetric()
	} else {
		required = requiredFieldsAsymmetric()
	}

	fields := map[string]string{
		"SecretKey":     rawConfig.SecretKey,
		"PrivateKeyPEM": rawConfig.PrivateKeyPEM,
		"PublicKeyPEM":  rawConfig.PublicKeyPEM,
	}

	for fieldName := range required {
		if fields[fieldName] == "" {
			return fmt.Errorf("%s required for %s algorithm", fieldName, rawConfig.Algorithm)
		}
	}

	return nil
}

// RawConfigToConfig parses RawConfig into Config with actual crypto keys
func RawConfigToConfig(rawConfig RawConfig) (*Config, error) {
	err := ValidateRawConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to validate raw config: %w", err)
	}

	algorithm, ok := Algorithms[rawConfig.Algorithm]
	if !ok {
		return nil, fmt.Errorf("no parser for algorithm: %s", rawConfig.Algorithm)
	}

	return algorithm.ParseRawConfig(rawConfig)
}
