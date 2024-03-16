package main

import (
	db "backend/db/sqlc"
	configutil "backend/util/config"
	logutil "backend/util/log"
	"context"
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var version string

func main() {
	logutil.GetLogger().Info("init-db, version: ", version)

	config, err := configutil.Load(os.Getenv("ENV"))
	if err != nil {
		logutil.GetLogger().Fatalf("load config error, err=%s", err)
	}
	logutil.GetLogger().Infof("configFile=%s", configutil.GetConfigFile())

	logutil.GetLogger().Infof("init db connection, source=%s", config.DB.Source)
	conn, err := sql.Open("pgx", config.DB.Source)
	if err != nil {
		logutil.GetLogger().Fatalf("init db connection error, err=%s", err)
	}
	if err := conn.Ping(); err != nil {
		logutil.GetLogger().Fatalf("init db connection error, err=%s", err)
	}

	store := db.NewStore(conn)

	if err := store.InitDB(context.Background()); err != nil {
		logutil.GetLogger().Fatalf("init db error, err=%s", err)
	}
}
