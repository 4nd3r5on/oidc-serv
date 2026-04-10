package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/4nd3r5on/oidc-serv/internal/keymanager"
)

type unmarshalFunc func(data []byte, v any) error

var unmarshalers = map[string]unmarshalFunc{
	".yml":  yaml.Unmarshal,
	".yaml": yaml.Unmarshal,
	".json": json.Unmarshal,
}

func getSupportedUnmarshalExts() []string {
	exts := make([]string, 0, len(unmarshalers))
	for ext := range unmarshalers {
		exts = append(exts, ext)
	}
	return exts
}

func readConfig(cfgPath string) (cfg keymanager.FileConfig, err error) {
	ext := path.Ext(cfgPath)
	unmarshal, ok := unmarshalers[ext]
	if !ok {
		return cfg, fmt.Errorf(
			"not supported JWT config file extension %q, expected %s",
			ext, strings.Join(getSupportedUnmarshalExts(), ", "),
		)
	}

	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read the file: %v", err)
	}

	err = unmarshal(cfgData, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config file: %v", err)
	}
	return cfg, nil
}

func resolveJWTConfig(cfgPath string, envPrefix string) (*keymanager.Config, error) {
	fileCfg, err := readConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("loading JWT config file %q: %v", cfgPath, err)
	}
	cfg, err := keymanager.ResolveConfig(fileCfg, envPrefix, slog.Default())
	if err != nil {
		return nil, fmt.Errorf("resolving JWT config: %v", err)
	}
	return cfg, nil
}
