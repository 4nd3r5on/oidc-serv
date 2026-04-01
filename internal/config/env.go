// Package config provides configuration-related helpers and definitions
package config

import (
	"log/slog"
	"os"
	"strings"
)

type EnvirnomentData struct {
	Name     string
	LogLevel slog.Level
}

const (
	// required
	// prod, dev, test
	EnvEnvironment = "ENVIRONMENT"
	// optional
	// :8080
	// 0.0.0.0:8080
	EnvServerAddr = "SERVER_ADDR"
)

type Environment int8

const (
	EnvironmentUnknown Environment = iota
	EnvironmentProd
	EnvironmentDev
	EnvironmentTest
)

var StrEnvirnomentMap = map[string]Environment{
	"prod": EnvironmentProd,
	"dev":  EnvironmentDev,
	"test": EnvironmentTest,
}

var EnvirnomentMap = map[Environment]EnvirnomentData{
	EnvironmentProd: {
		Name:     "prod",
		LogLevel: slog.LevelInfo,
	},
	EnvironmentDev: {
		Name:     "dev",
		LogLevel: slog.LevelDebug,
	},
	EnvironmentTest: {
		Name:     "test",
		LogLevel: slog.LevelDebug,
	},
}

func StrToEnvirnoment(envStr string) Environment {
	env, ok := StrEnvirnomentMap[envStr]
	if !ok {
		return EnvironmentUnknown
	}
	return env
}

func GetEnvironment() Environment {
	return StrToEnvirnoment(
		strings.ToLower(os.Getenv(EnvEnvironment)))
}

func (env Environment) LogLevel() slog.Level {
	data, ok := EnvirnomentMap[env]
	if !ok {
		return slog.LevelDebug
	}
	return data.LogLevel
}

func (env Environment) String() string {
	data, ok := EnvirnomentMap[env]
	if !ok {
		return ""
	}
	return data.Name
}
