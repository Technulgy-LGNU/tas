package util

import (
	"log"
	"os"
)

func CheckDirs() {
	// data dir
	if _, err := os.Stat("data/"); os.IsNotExist(err) {
		_, err := os.Create("data/")
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data dir: %v", err)
		}
	}

	// log dir
	if _, err := os.Stat("data/gorm_logs"); os.IsNotExist(err) {
		_, err := os.Create("data/gorm_logs")
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data/gorm_logs dir: %v", err)
		}
	}
	if _, err := os.Stat("data/fiber_logs"); os.IsNotExist(err) {
		_, err := os.Create("data/fiber_logs")
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data dir: %v", err)
		}
	}

	// cdn dir
	if _, err := os.Stat("data/cdn"); os.IsNotExist(err) {
		_, err := os.Create("data/cdn")
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data/cdn dir: %v", err)
		}
	}
}
