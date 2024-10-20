package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tas/src/config"
	"tas/src/database"
	cLog "tas/src/log"
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
	log.Println("Starting T.A.S. ...")

	// CheckDirs (If not, error on log write)
	util.CheckDirs()

	// Config
	var CFG = config.GetConfig()

	// GormLogger
	gormLogger := &cLog.GormCustomLogger{}
	gormLogger.StartDailyFlush()
	// Database
	DB, err := database.GetDatabase(gormLogger, CFG)
	if err != nil {
		log.Fatal(err)
	}
	// Do the initial checks parallel to save start up time
	go func() {
		err = database.InitDatabase(CFG, DB)
		if err != nil {
			log.Fatal(err)
		}
		// Takes the longest to finish, so total startup time is measured here
		mst.ElapsedTime()
	}()

	// Util

	// Routines
	util.DeleteOldSessions(DB)
	util.DeleteSoftDeletedUserKeys(DB)

	// FiberLogger
	fiberLogger := &cLog.FiberCustomLogger{}
	fiberLogger.StartDailyFlush()
	// Web
	err = web.InitWeb(fiberLogger, CFG, DB)
	if err != nil {
		log.Fatal(err)
	}

	// Handle shutdown (Not working, don't know why ...)
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGKILL)

	sig := <-sigs
	log.Println()
	log.Printf("Caught signal %s; exiting...", sig)

	gormLogger.WriteLogToDisk()
	fiberLogger.WriteLogToDisk()

	done <- true

	<-done
	log.Println("Shutting down...")
	os.Exit(0)
}
