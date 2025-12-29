package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port                int    `mapstructure:"port"`
		PublicURL           string `mapstructure:"public_url"`
		TelegramWebhookPath string `mapstructure:"telegram_webhook_path"`
		GithubWebhookPath   string `mapstructure:"github_webhook_path"`
		TLS                 struct {
			Enabled  bool   `mapstructure:"enabled"`
			CertFile string `mapstructure:"cert_file"`
			KeyFile  string `mapstructure:"key_file"`
		} `mapstructure:"tls"`
	} `mapstructure:"server"`

	Telegram struct {
		BotToken string `mapstructure:"bot_token"`
	} `mapstructure:"telegram"`

	Github struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"github"`

	Log struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`

	DB struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"db"`
}

func Load() (Config, error) {
	v := viper.New()

	v.SetConfigFile("config/config.yml")
	v.SetConfigType("yaml")

	v.SetDefault("server.port", 8080)
	v.SetDefault("server.telegram_webhook_path", "/api/v1/telegram/webhook")
	v.SetDefault("server.github_webhook_path", "/api/v1/github/webhook")
	v.SetDefault("log.level", "info")

	_ = v.ReadInConfig()

	v.SetEnvPrefix("CRNB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.BindEnv("telegram.bot_token", "CRNB_TELEGRAM_BOT_TOKEN"); err != nil {
		return Config{}, fmt.Errorf("bind env CRNB_TELEGRAM_BOT_TOKEN: %w", err)
	}
	if err := v.BindEnv("github.secret", "CRNB_GITHUB_SECRET"); err != nil {
		return Config{}, fmt.Errorf("bind env CRNB_GITHUB_SECRET: %w", err)
	}
	if err := v.BindEnv("server.public_url", "CRNB_SERVER_PUBLIC_URL"); err != nil {
		return Config{}, fmt.Errorf("bind env CRNB_SERVER_PUBLIC_URL: %w", err)
	}
	if err := v.BindEnv("db.dsn", "CRNB_DB_DSN"); err != nil {
		return Config{}, fmt.Errorf("bind env CRNB_DB_DSN: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
