package util

import (
	"log"
	"os"
)

func CheckDirs() {
	// data dir
	if _, err := os.Stat("data/"); os.IsNotExist(err) {
		err := os.Mkdir("data", 0777)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data dir: %v", err)
		}
	}

	// log dir
	if _, err := os.Stat("data/gorm_logs"); os.IsNotExist(err) {
		err := os.Mkdir("data/gorm_logs", 0777)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data/gorm_logs dir: %v", err)
		}
	}
	if _, err := os.Stat("data/fiber_logs"); os.IsNotExist(err) {
		err := os.Mkdir("data/fiber_logs", 0777)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data dir: %v", err)
		}
	}

	// cdn dir
	if _, err := os.Stat("data/cdn"); os.IsNotExist(err) {
		err := os.Mkdir("data/cdn", 0777)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error creating data/cdn dir: %v", err)
		}
	}
}
