package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"tas/backend/config"
	"tas/backend/database"
	"tas/backend/observability"
	"tas/backend/web"
)

func main() {
	// Used by the container build to check the executable before publication.
	if len(os.Args) == 2 && os.Args[1] == "--build-info" {
		fmt.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}
	var cfg = config.GetConfig()
	if err := observability.Configure(cfg.Logging.Level); err != nil {
		log.Fatal(err)
	}
	slog.Info("startup.database_connecting")

	var db = database.GetDB(cfg)
	slog.Info("startup.database_connected")
	database.InitDB(db)
	slog.Info("startup.migrations_complete")

	web.InitWeb(cfg, db)
}
