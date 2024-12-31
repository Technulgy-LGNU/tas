package util

import (
	"os"
	clog "tas/src/log"
)

// CheckDirs checks if the necessary directories exist and creates them if they don't
func CheckDirs(logger *clog.Logger) {
	if os.Args[1] == "prod" {
		// data dir
		if _, err := os.Stat("/var/lib/tas"); os.IsNotExist(err) {
			err := os.MkdirAll("/var/lib/tas", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}

		// log dir
		if _, err := os.Stat("/var/lib/tas/logs"); os.IsNotExist(err) {
			err := os.MkdirAll("/var/lib/tas/logs", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}

		// cdn dir
		if _, err := os.Stat("/var/lib/tas/cdn"); os.IsNotExist(err) {
			err := os.MkdirAll("/var/lib/tas/cdn", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}
	} else if os.Args[1] == "dev" {
		// data dir
		if _, err := os.Stat("data/"); os.IsNotExist(err) {
			err := os.MkdirAll("data", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}

		// log dir
		if _, err := os.Stat("data/logs"); os.IsNotExist(err) {
			err := os.MkdirAll("data/logs", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}

		// cdn dir
		if _, err := os.Stat("data/cdn"); os.IsNotExist(err) {
			err := os.MkdirAll("data/cdn", 0777)
			if err != nil {
				logger.LogEvent(err.Error(), "FATAL")
			}
		}
	}
}
