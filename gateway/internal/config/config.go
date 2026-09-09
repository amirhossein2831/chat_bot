package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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
	Env       string          `mapstructure:"env"      validate:"required,oneof=local staging production"`
	HTTP      HTTPConfig      `mapstructure:"http"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Relay     RelayConfig     `mapstructure:"relay"`
	Models    []ModelConfig   `mapstructure:"models"`
	Objects   ObjectsConfig   `mapstructure:"objects"`
	Files     FilesConfig     `mapstructure:"files"`
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

type ObjectsConfig struct {
	Endpoint  string `mapstructure:"endpoint"   validate:"required"`
	AccessKey string `mapstructure:"access_key" validate:"required"`
	SecretKey string `mapstructure:"secret_key" validate:"required"`
	Bucket    string `mapstructure:"bucket"     validate:"required"`
	UseTLS    bool   `mapstructure:"use_tls"`
}

type FilesConfig struct {
	MaxBytes      int64 `mapstructure:"max_bytes"       validate:"min=1"`
	PresignTTLMin int   `mapstructure:"presign_ttl_min" validate:"min=1"`
	DailyLimit    int   `mapstructure:"daily_limit"     validate:"min=1"`
}

func (f FilesConfig) PresignTTL() time.Duration {
	return time.Duration(f.PresignTTLMin) * time.Minute
}

type ProvidersConfig struct {
	Local     ProviderConfig `mapstructure:"local"`
	OpenAI    ProviderConfig `mapstructure:"openai"`
	Anthropic ProviderConfig `mapstructure:"anthropic"`
}

type ProviderConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}

type RelayConfig struct {
	Endpoints  []string `mapstructure:"endpoints"`
	Credential string   `mapstructure:"credential"`
}

type ModelConfig struct {
	ID                    string `mapstructure:"id"                          validate:"required"`
	Provider              string `mapstructure:"provider"                    validate:"required,oneof=local openai anthropic"`
	Tier                  string `mapstructure:"tier"                        validate:"required,oneof=local frontier"`
	ContextWindow         int    `mapstructure:"context_window"              validate:"min=1"`
	CostPer1kInUSDMicros  int64  `mapstructure:"cost_per_1k_in_usd_micros"`
	CostPer1kOutUSDMicros int64  `mapstructure:"cost_per_1k_out_usd_micros"`
}

type AuthConfig struct {
	SessionTTLHours int `mapstructure:"session_ttl_hours" validate:"min=1"`
}

func (a AuthConfig) SessionTTL() time.Duration {
	return time.Duration(a.SessionTTLHours) * time.Hour
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

	v.SetDefault("auth.session_ttl_hours", 720)

	v.SetDefault("providers.local.base_url", "")
	v.SetDefault("providers.local.api_key", "")
	v.SetDefault("providers.openai.base_url", "https://api.openai.com")
	v.SetDefault("providers.openai.api_key", "")
	v.SetDefault("providers.anthropic.base_url", "https://api.anthropic.com")
	v.SetDefault("providers.anthropic.api_key", "")

	v.SetDefault("relay.endpoints", []string{})
	v.SetDefault("relay.credential", "")

	v.SetDefault("models", []ModelConfig{})

	v.SetDefault("objects.endpoint", "localhost:9000")
	v.SetDefault("objects.access_key", "minioadmin")
	v.SetDefault("objects.secret_key", "minioadmin")
	v.SetDefault("objects.bucket", "attachments")
	v.SetDefault("objects.use_tls", false)

	v.SetDefault("files.max_bytes", 20*1024*1024)
	v.SetDefault("files.presign_ttl_min", 15)
	v.SetDefault("files.daily_limit", 20)
}
