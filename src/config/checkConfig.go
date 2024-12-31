package config

import (
	"gopkg.in/yaml.v3"
	"os"
	clog "tas/src/log"
)

// CheckConfig
// This function checks the config file for errors, if there are any, it will panic
// It also checks if the config file exists and if not create one
func CheckConfig(logger *clog.Logger) {
	// Should be self-explanatory, if not, open a GitHub Ticket with the questions tag
	if os.Args[1] == "prod" {
		if _, err := os.Stat("/var/lib/tas/config/config.yaml"); os.IsNotExist(err) {
			logger.LogEvent("Config file not found, creating one", "INFO")
			err := os.MkdirAll("/var/lib/tas/config", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
			_, err = os.Create("/var/lib/tas/config/config.yaml")
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}
	} else if os.Args[1] == "dev" {
		if _, err := os.Stat("config/config.yaml"); os.IsNotExist(err) {
			// Create the config file
			logger.LogEvent("Config file not found, creating one", "INFO")
			err := os.MkdirAll("config", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
			_, err = os.Create("config/config.yaml")
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}

			// Write the default config to the file
			file, err := os.OpenFile("config/config.yaml", os.O_WRONLY, 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
			defer file.Close()

			data, err := yaml.Marshal(CFG{})
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}

			_, err = file.Write(data)
		}
	} else {
		panic("Error: Wrong command line argument")
	}
}
