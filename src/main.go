package main

import (
	"log"
	"tas/src/config"
	"tas/src/database"
	cLog "tas/src/log"
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

	// GormLogger
	gormLogger := &cLog.GormCustomLogger{}
	gormLogger.StartDailyFlush()
	gormLogger.HandleShutdown()
	// Database
	var DB = database.GetDatabase(gormLogger, CFG)
	database.InitDatabase(CFG, DB)

	// Util

	// Routines

	// FiberLogger
	fiberLogger := &cLog.GormCustomLogger{}
	fiberLogger.StartDailyFlush()
	fiberLogger.HandleShutdown()
	// Web
	web.InitWeb((*cLog.FiberCustomLogger)(fiberLogger), CFG, DB)
}
