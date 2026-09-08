// Package config loads and validates application configuration.
//
// This is the only package that imports viper. Everything downstream receives
// a typed *Config, so a renamed key breaks compilation instead of production.
//
// Precedence, lowest to highest:
//
//	code defaults  ->  config.yaml  ->  config.<env>.yaml  ->  APP_* env vars
//
// The environment is chosen by APP_ENV and defaults to "local". Secrets belong
// in the environment, never in a committed per-environment file.
//
// # Environment variable naming
//
// Any key can be overridden by an environment variable named APP_ plus the
// key's path in upper case with every "." replaced by "_":
//
//	env                     ->  APP_ENV
//	http.addr               ->  APP_HTTP_ADDR
//	http.shutdown_timeout   ->  APP_HTTP_SHUTDOWN_TIMEOUT
//	postgres.dsn            ->  APP_POSTGRES_DSN
//	postgres.max_conns      ->  APP_POSTGRES_MAX_CONNS
//	redis.addr              ->  APP_REDIS_ADDR
//	redis.db                ->  APP_REDIS_DB
//
// A key that is not registered in setDefaults is never consulted in the
// environment, so adding a field to Config also requires adding its default.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	EnvLocal      = "local"
	EnvStaging    = "staging"
	EnvProduction = "production"

	defaultConfigDir = "./configs"
)

type Config struct {
	Env      string         `mapstructure:"env"      validate:"required,oneof=local staging production"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

func (c Config) IsProduction() bool { return c.Env == EnvProduction }

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

// Load reads configuration from the directory named by APP_CONFIG_DIR, or
// ./configs when that is unset.
func Load() (*Config, error) {
	dir := os.Getenv("APP_CONFIG_DIR")
	if dir == "" {
		dir = defaultConfigDir
	}
	return LoadFrom(dir)
}

// LoadFrom reads configuration from an explicit directory. Tests use this to
// point at a temporary directory instead of the repository's configs/.
func LoadFrom(dir string) (*Config, error) {
	env := environment()

	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v, env)

	// Base file, then the environment overlay. Both are optional: a container
	// configured entirely through APP_* variables is a valid deployment.
	v.SetConfigName("config")
	if err := readOptional(v.ReadInConfig); err != nil {
		return nil, fmt.Errorf("reading config.yaml: %w", err)
	}

	v.SetConfigName("config." + env)
	if err := readOptional(v.MergeInConfig); err != nil {
		return nil, fmt.Errorf("reading config.%s.yaml: %w", env, err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	if err := validator.New().Struct(&c); err != nil {
		return nil, fmt.Errorf("invalid configuration for env %q: %w", env, err)
	}
	return &c, nil
}

func readOptional(read func() error) error {
	err := read()
	if err == nil {
		return nil
	}
	var notFound viper.ConfigFileNotFoundError
	if errors.As(err, &notFound) {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// environment is read straight from the process environment rather than
// through viper, because it selects which file viper reads.
func environment() string {
	if e := strings.TrimSpace(os.Getenv("APP_ENV")); e != "" {
		return e
	}
	return EnvLocal
}

// setDefaults registers every key the application reads.
//
// This is not only about defaults: viper's Unmarshal walks the keys it knows
// about, and a key registered nowhere is never consulted in the environment.
// Required values are registered with an empty default so that APP_* overrides
// are picked up and validation catches them when they are absent.
func setDefaults(v *viper.Viper, env string) {
	v.SetDefault("env", env)

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.shutdown_timeout", 30)

	v.SetDefault("postgres.dsn", "")
	v.SetDefault("postgres.max_conns", 10)

	v.SetDefault("redis.addr", "")
	v.SetDefault("redis.db", 0)
}
