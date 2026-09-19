package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
  Database struct {
    Host string `toml:"host"`
    Port int `toml:"port"`
    User string `toml:"user"`
    Password string `toml:"password"`
    Database string `toml:"database"`
    TimeZone string `toml:"timezone"`
  } `toml:"database"`

  Auth struct {
    FusionAuthURL string `toml:"fusionauht_url"`
    FusionAuthClientId string `toml:"fusionauht_client_id"`
    FusionAuthSecret string `toml:"fusionauht_fusionauth_secret"`
    FusionAuthTenantId string `toml:"fusionauht_tenant_id"`
    OAuthRedirectURI string `toml:"oauth_redirect_uri"`
    FrontendURL string `toml:"frontend_url"`
  } `toml:"auth"`

  Cloudflare struct {
    ImagesAccountId string `toml:"images_account_id"`
    ImagesAPIToken string `toml:"images_api_token"`
    ImagesDeliveryURL string `toml:"images_delivery_url"`
    ImagesVariant string `toml:"images_variant"`
  } `toml:"cloudflare"`
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
