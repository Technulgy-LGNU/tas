package main

import (
	"log"
	"tas/src/config"
	"tas/src/database"
	clog "tas/src/log"
	"tas/src/util"
	"tas/src/web"
)

// Welcome to T.A.S. (Technulgy Admin Software)
// This software is for managing members, teams, sponsors, orders and events
// Currently in Development, look under projects to see the current state

func main() {
	// Logger
	logger := clog.Logger{}
	gormLogger := clog.GormLogger{
		L: &logger,
	}
	fiberLogger := clog.FiberLogger{
		L: &logger,
	}

	// Start timer
	var mst util.MST
	mst.StartTimer()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting TAS ...")

	// Checks
	config.CheckConfig(&logger)

	// Config
	var CFG = config.GetConfig()

	// Database
	DB, err := database.GetDatabase(&gormLogger, CFG)
	if err != nil {
		logger.LogEvent(err.Error(), "FATAL")
	}
	// Do the initial checks parallel to save start up time
	go func() {
		err = database.InitDatabase(&logger, DB)
		if err != nil {
			logger.LogEvent(err.Error(), "FATAL")
		}
		// Takes the longest to finish, so total startup time is measured here
		mst.ElapsedTime()
	}()

	// Routines
	util.DeleteOldSessions(DB)
	util.DeleteSoftDeletedUserKeys(DB)

	// Web
	err = web.InitWeb(&fiberLogger, &logger, CFG, DB)
	if err != nil {
		logger.LogEvent(err.Error(), "FATAL")
	}
}
