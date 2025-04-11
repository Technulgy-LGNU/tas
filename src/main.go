package main

import (
	"errors"
	"gorm.io/gorm"
	"log"
	"tas/src/config"
	"tas/src/database"
	"tas/src/util"
	"tas/src/web"
)

// Welcome to T.A.S. (Technulgy Admin Software)
// This software is for managing members, teams, sponsors, orders and events
// Currently in Development, look under projects to see the current state

func main() {
	// Start timer
	var mst util.MST
	mst.StartTimer()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting TAS ...")

	// Checks

	// Config
	var CFG = config.GetConfig()

	// Database
	DB, err := database.GetDatabase(CFG)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	// Do the initial checks parallel to save start up time
	go func() {
		err = database.InitDatabase(DB)
		if err != nil {
			log.Fatalf("Error initializing database: %v", err)
		}

		// Create Team Null, if it doesn't exist
		var team database.Team
		result := DB.First(&team, 1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Println("Team Null not found, creating...")
				var teamNull = database.Team{
					Model:   gorm.Model{},
					Name:    "noTeam",
					League:  "",
					Members: nil,
					History: nil,
					EventID: nil,
					Event:   nil,
				}
				err := DB.Create(&teamNull).Error
				if err != nil {
					log.Fatalf("Error creating team: %v", err)
				}
			} else {
				log.Fatalf("Error finding team: %v", result.Error)
			}
		}

		// Takes the longest to finish, so total startup time is measured here
		mst.ElapsedTime()
	}()

	// Routines
	util.DeleteOldSessions(DB)
	util.DeleteSoftDeletedUserKeys(DB)
	util.DeleteOldTDPs(DB)
	util.DeleteOldForms(DB)
	util.DeleteOldPasswordResetCodes(DB)

	// Web
	web.InitWeb(CFG, DB)
}
