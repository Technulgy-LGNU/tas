package database

import (
	"fmt"
	"tas/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetPSQL(cfg *config.Config) (*gorm.DB, error) {
	var (
		dbURI = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s TimeZone=%s",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Database,
			cfg.Database.Pass,
			cfg.Database.TimeZone,
		)
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, nil
}
