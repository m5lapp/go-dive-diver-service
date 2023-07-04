package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/m5lapp/go-dive-diver-service/internal/data"
	"github.com/m5lapp/go-service-toolkit/config"
	"github.com/m5lapp/go-service-toolkit/persistence/sqldb"
	"github.com/m5lapp/go-service-toolkit/vcs"
	"github.com/m5lapp/go-service-toolkit/webapp"
	"golang.org/x/exp/slog"
)

type appConfig struct {
	db      config.SqlDB
	svcUser config.Service
}

type app struct {
	webapp.WebApp
	cfg    appConfig
	models data.Models
}

func main() {
	var serverCfg config.Server
	var appCfg appConfig

	serverCfg.Flags(":8080")
	appCfg.db.Flags("postgres", 25, 25, "15m")
	appCfg.svcUser.Flags("user-service-address", "HTTP address of the user service")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", vcs.Version())
		os.Exit(0)
	}

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(logHandler)
	db, err := sqldb.OpenDB(appCfg.db)
	if err != nil {
		logger.Error(err.Error(), nil)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database connection pool established")

	app := &app{
		WebApp: webapp.New(serverCfg, logger),
		cfg:    appCfg,
		models: data.NewModels(db),
	}

	err = app.Serve(app.routes())
	if err != nil {
		logger.Error(err.Error(), nil)
		os.Exit(1)
	}
}
