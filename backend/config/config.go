package config

import (
	"log"
	"os"
	"tas/backend/contact"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Logging struct {
		Level string `toml:"level"`
	} `toml:"logging"`
	Website struct {
		AllowedOrigins []string       `toml:"allowed_origins"`
		PublicURL      string         `toml:"public_url"`
		Contact        contact.Config `toml:"contact"`
	} `toml:"website"`
	Database struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
		Database string `toml:"database"`
		TimeZone string `toml:"timezone"`
	} `toml:"database"`

	Auth AuthConfig `toml:"auth"`

	Cloudflare struct {
		ImagesAccountId       string `toml:"images_account_id"`
		ImagesAPIToken        string `toml:"images_api_token"`
		ImagesDeliveryURL     string `toml:"images_delivery_url"`
		ImagesVariant         string `toml:"images_variant"`
		ImagesTransformOrigin string `toml:"images_transform_origin"`
	} `toml:"cloudflare"`
}

type AuthConfig struct {
	DisableFusionAuth  bool   `toml:"disable_fusionauth"`
	FusionAuthURL      string `toml:"fusionauth_url"`
	FusionAuthClientId string `toml:"fusionauth_client_id"`
	FusionAuthSecret   string `toml:"fusionauth_client_secret"`
	FusionAuthTenantId string `toml:"fusionauth_tenant_id"`
	OAuthRedirectURI   string `toml:"oauth_redirect_uri"`
	FrontendURL        string `toml:"frontend_url"`
	RequiredRole       string `toml:"required_role"`
}

func GetConfig() *Config {
	var cfg Config

	bytes, err := os.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("Error opening config file: %v\n", err)
	}

	if err := toml.Unmarshal(bytes, &cfg); err != nil {
		log.Fatalf("Error unmarshaling config file: %v\n", err)
	}

	return &cfg
}
