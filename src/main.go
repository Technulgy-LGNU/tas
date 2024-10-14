package main

import (
	"log"
	"tas/src/config"
	"tas/src/database"
	"tas/src/web"
)

// Welcome to T.A.S. (Technulgy Admin Software)
// This software is for managing members, teams, sponsors, orders and events
// Currently in Development, look under projects to see the current state

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting T.A.S. ...")

	// Config
	var CFG = config.GetConfig()

	// Database
	var DB = database.GetDatabase(CFG)
	database.InitDatabase(DB)

	// Util

	// Routines

	// Web
	web.InitWeb(CFG, DB)
}
