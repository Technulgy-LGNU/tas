package database

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"tas/src/config"
	cLog "tas/src/log"
)

func GetDatabase(customLogger *cLog.GormCustomLogger, cfg *config.CFG) (*gorm.DB, error) {
	var (
		dbURI = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s TimeZone=%s",
			cfg.DB.Host,
			cfg.DB.Username,
			cfg.DB.Database,
			cfg.DB.Password,
			cfg.DB.TimeZone)
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error connecting to database: %v\n", err))
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Successfully connected to database")
	}

	return db, nil
}
