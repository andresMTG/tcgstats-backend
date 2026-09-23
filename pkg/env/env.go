package env

import "github.com/caarlos0/env/v9"

// LoadEnvConfiguration load configuration with caarlos0/env library.
func LoadEnvConfiguration[T any]() (*T, error) {
	var config T
	return &config, env.Parse(&config)
}