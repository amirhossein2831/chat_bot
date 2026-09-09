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

func Load() (*Config, error) {
	dir := os.Getenv("APP_CONFIG_DIR")
	if dir == "" {
		dir = defaultConfigDir
	}
	return LoadFrom(dir)
}

func LoadFrom(dir string) (*Config, error) {
	env := environment()

	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v, env)

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
	if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func environment() string {
	if e := strings.TrimSpace(os.Getenv("APP_ENV")); e != "" {
		return e
	}
	return EnvLocal
}

// Every key needs an entry here, including required ones with an empty value:
// viper never consults the environment for a key it has not seen, so a missing
// entry silently ignores its APP_* override.
func setDefaults(v *viper.Viper, env string) {
	v.SetDefault("env", env)

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.shutdown_timeout", 30)

	v.SetDefault("postgres.dsn", "")
	v.SetDefault("postgres.max_conns", 10)

	v.SetDefault("redis.addr", "")
	v.SetDefault("redis.db", 0)
}
