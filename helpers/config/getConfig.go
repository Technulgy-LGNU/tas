package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

type CFG struct {
	DB struct {
		Host     string `yaml:"Host"`
		Port     int    `yaml:"Port"`
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
		Database string `yaml:"Database"`
		TimeZone string `yaml:"TimeZone"`
	} `yaml:"Database"`

	Email struct {
		Host                string `yaml:"Host"`
		ApiKey              string `yaml:"ApiKey"`
		SenderEmail         string `yaml:"SenderEmail"`
		SenderEmailPassword string `yaml:"Password"`
	} `yaml:"Email"`

	Nextcloud struct {
		Host   string `yaml:"Host"`
		APIKey string `yaml:"ApiKey"`
	} `yaml:"Nextcloud"`

	DiscordWebhook string `yaml:"DiscordWebhook"`

	TDPUploadKey string `yaml:"TDPUpload_Key"`
}

// GetConfig
// For dev purposes, you can run the app as a compiled go file, but this setup needs additional options
// these can be configured in the config.yaml.
// In the production setup, these additional values are hardcoded in the program
func GetConfig() *CFG {
	// I think, this is all self-explanatory, so no further comments, on questions open a GitHub Ticket with the questions tag
	var (
		file    string
		cfgFile *os.File
		config  CFG

		err error
	)

	if os.Args[1] == "prod" {
		config.DB.Host = os.Getenv("DB_HOST")
		config.DB.Port, _ = strconv.Atoi(os.Getenv("DB_PORT"))
		config.DB.Username = os.Getenv("DB_USERNAME")
		config.DB.Password = os.Getenv("DB_PASSWORD")
		config.DB.Database = os.Getenv("DB_DATABASE")
		config.DB.TimeZone = os.Getenv("DB_TIMEZONE")
		config.Email.Host = os.Getenv("Email_HOST")
		config.Email.ApiKey = os.Getenv("EMAIL_API_KEY")
		config.Email.SenderEmail = os.Getenv("EMAIL_SENDER_EMAIL")
		config.Email.SenderEmailPassword = os.Getenv("EMAIL_SENDER_EMAIL_PASSWORD")
		config.Nextcloud.Host = os.Getenv("NEXTCLOUD_HOST")
		config.Nextcloud.APIKey = os.Getenv("NEXTCLOUD_API_KEY")
		config.DiscordWebhook = os.Getenv("DISCORD_WEBHOOK")
		config.TDPUploadKey = os.Getenv("TDPUpload_Key")

		return &config
	} else if os.Args[1] == "dev" {
		file = "../config/config.yaml"

		cfgFile, err = os.Open(file)
		if err != nil {
			log.SetFlags(log.LstdFlags & log.Lshortfile)
			log.Fatalf("Error readeing config file: %d\n", err)
		}

		yamlParser := yaml.NewDecoder(cfgFile)
		err = yamlParser.Decode(&config)
		if err != nil {
			log.SetFlags(log.LstdFlags & log.Lshortfile)
			log.Fatalf("Error readeing config file: %d\n", err)
		}

		return &config
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatalf("Error: No config file found, please run the program with 'prod' or 'dev' as argument")
		return nil
	}
}
