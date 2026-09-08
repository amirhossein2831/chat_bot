// Package config loads and validates application configuration.
//
// This is the only package that imports viper. Everything downstream receives
// a typed *Config, so a renamed key breaks compilation instead of production.
package config

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Env      string         `mapstructure:"env"      validate:"required,oneof=dev staging prod"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type HTTPConfig struct {
	Addr            string `mapstructure:"addr"             validate:"required"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout" validate:"min=1"`
}

type PostgresConfig struct {
	DSN      string `mapstructure:"dsn"       validate:"required"`
	MaxConns int32  `mapstructure:"max_conns" validate:"min=1"`
}

type RedisConfig struct {
	Addr string `mapstructure:"addr" validate:"required"`
	DB   int    `mapstructure:"db"   validate:"min=0"`
}

// Load reads configs/config.yaml if present, overlays APP_* environment
// variables, and validates the result.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/configs")
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
		// A missing file is fine: environment-only configuration is valid.
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	if err := validator.New().Struct(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// setDefaults registers every key the application reads.
//
// This is not only about defaults: viper's Unmarshal walks the keys it knows
// about, and a key registered nowhere is never consulted in the environment.
// Required values are registered with an empty default so that APP_* overrides
// are picked up and validation catches them when they are absent.
func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "dev")

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.shutdown_timeout", 30)

	v.SetDefault("postgres.dsn", "")
	v.SetDefault("postgres.max_conns", 10)

	v.SetDefault("redis.addr", "")
	v.SetDefault("redis.db", 0)
}
