package database

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tas/src/config"
)

func GetDatabase(cfg *config.CFG) (*gorm.DB, error) {
	var (
		dbURI = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s TimeZone=%s",
			cfg.DB.Host,
			cfg.DB.Username,
			cfg.DB.Database,
			cfg.DB.Password,
			cfg.DB.TimeZone,
		)
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error connecting to database: %v\n", err))
	} else {
	}

	return db, nil
}
