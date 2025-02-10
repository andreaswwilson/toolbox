package config

import (
	"context"
)

// Define a type for the context key
type ContextKey string

const ConfigKey ContextKey = "configData"

// WithConfig adds the config to the context
func WithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, ConfigKey, config)
}

// FromContext retrieves the config from the context
func FromContext(ctx context.Context) (*Config, bool) {
	config, ok := ctx.Value(ConfigKey).(*Config)
	return config, ok
}
