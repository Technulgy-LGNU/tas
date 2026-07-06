package main

import (
	"log"
	"tas/internal/config"
	"tas/internal/database"
	"tas/internal/web"
)

func main() {
	// Starting T.A.S. (Technulgy Admin Software)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting T.A.S. (Technulgy Admin Software)")

	cfg := config.GetConfig()
	db, err := database.GetPSQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.InitPSQL(db); err != nil {
		log.Fatal(err)
	}

	web.InitWeb(cfg, db)
}
