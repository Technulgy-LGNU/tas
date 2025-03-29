package database

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"tas/src/config"
	cLog "tas/src/log"
)

func GetDatabase(log *cLog.GormLogger, cfg *config.CFG) (*gorm.DB, error) {
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
		Logger: log,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error connecting to database: %v\n", err))
	} else {
	}

	return db, nil
}
