package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Database struct {
		Host     string `toml:"host"`
		Port     uint16 `toml:"port"`
		User     string `toml:"user"`
		Pass     string `toml:"pass"`
		Database string `toml:"database"`
		TimeZone string `toml:"timezone"`
	} `toml:"database"`

	Server struct {
		Host string `toml:"host"`
		Port uint16 `toml:"port"`
	} `toml:"server"`

	Auth struct {
		Issuer        string `toml:"issuer"`
		ClientID      string `toml:"client_id"`
		RequiredRole  string `toml:"required_role"`
		DevAllowAdmin bool   `toml:"dev_allow_admin"`
	} `toml:"auth"`
}

func defaultConfig() Config {
	var cfg Config

	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "rcjv_paperless"
	cfg.Database.Pass = "password"
	cfg.Database.Database = "rcjv_paperless"
	cfg.Database.TimeZone = "Europe/Berlin"

	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8000

	cfg.Auth.RequiredRole = "admin"

	return cfg
}

func normalizeBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	return strings.TrimRight(raw, "/") + "/"
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	bytes, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &cfg, nil
		}
		return nil, err
	}
	if err := toml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	cfg.Auth.Issuer = strings.TrimRight(strings.TrimSpace(cfg.Auth.Issuer), "/")

	return &cfg, nil
}

func GetConfig() *Config {
	cfg, err := Load("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
