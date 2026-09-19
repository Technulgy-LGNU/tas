package database

import (
	"fmt"
	"log"
	"tas/backend/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetDB(cfg *config.Config) *gorm.DB {
  var dbURI = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s",
    cfg.Database.Host,
    cfg.Database.Port,
    cfg.Database.User,
    cfg.Database.Database,
    cfg.Database.Password,
  )

  // Open connection to database
  db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
    DisableForeignKeyConstraintWhenMigrating: true,
    Logger: logger.Default.LogMode(logger.Error),
  })
  if err != nil {
    log.Fatalf("Error connecting to database: %v\n", err)
  }

  return db
}
