package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"tas/src/config"
)

func GetDatabase(cfg *config.CFG) *gorm.DB {
	var (
		dbURI = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s TimeZone=%s",
			cfg.DB.Host,
			cfg.DB.Username,
			cfg.DB.Database,
			cfg.DB.Password,
			cfg.DB.TimeZone)

		err error = nil
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatalf("Error connecting to database: %d\n", err)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Successfully connected to database")
	}

	return db
}
