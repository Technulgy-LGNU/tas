package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type CFG struct {
	LogLevel string `yaml:"LogLevel"`

	DB struct {
		Host     string `yaml:"Host"`
		Port     int    `yaml:"Port"`
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
		Database string `yaml:"Database"`
		TimeZone string `yaml:"TimeZone"`
	} `yaml:"Database"`

	User struct {
		AdminPassword string `yaml:"InitialAdminPassword"`
	} `yaml:"User"`

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
		file = "/var/lib/tas/config/config.yaml"
	} else if os.Args[1] == "dev" {
		file = "config/config.yaml"
	}

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
}
