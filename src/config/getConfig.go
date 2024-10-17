package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
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

	User struct {
		AdminUserName string `yaml:"InitialAdminUser"`
		AdminPassword string `yaml:"InitialAdminPassword"`
	} `yaml:"User"`

	Website struct {
		Host string `yaml:"Host"`
		Port int    `yaml:"Port"`
	} `yaml:"Website"`

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

func GetConfig() *CFG {
	if os.Args[1] == "prod" {
		var c CFG
		c.DB.Host = "db"
		c.DB.Port = 5432
		c.DB.Username = os.Getenv("DBUser")
		c.DB.Password = os.Getenv("DBPassword")
		c.DB.Database = os.Getenv("Database")
		c.DB.TimeZone = os.Getenv("TimeZone")
		c.User.AdminUserName = os.Getenv("InitialAdminUser")
		c.User.AdminPassword = os.Getenv("InitialAdminPassword")
		c.Website.Host = "0.0.0.0"
		c.Website.Port = 80
		c.Email.Host = os.Getenv("EmailHost")
		c.Email.ApiKey = os.Getenv("EmailApiKey")
		c.Email.SenderEmail = os.Getenv("EmailSenderEmail")
		c.Email.SenderEmailPassword = os.Getenv("EmailSenderEmailPassword")
		c.Nextcloud.Host = os.Getenv("NextcloudHost")
		c.Nextcloud.APIKey = os.Getenv("NextcloudApiKey")

		return &c
	} else if os.Args[1] == "dev" {
		const file = "config/config.yaml"
		var config CFG

		cfgFile, err := os.Open(file)
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
		panic("Error: Wrong command line argument")
		return nil
	}
}
