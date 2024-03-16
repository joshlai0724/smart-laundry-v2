package main

import (
	db "backend/db/sqlc"
	iotsdk "backend/iot-sdk"
	"backend/token"
	configutil "backend/util/config"
	logutil "backend/util/log"
	"backend/web"
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	_ "github.com/jackc/pgx/v4/stdlib"
	goredislib "github.com/redis/go-redis/v9"
)

var version string

func main() {
	logutil.GetLogger().Info("web-backend, version: ", version)

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

	logutil.GetLogger().Infof("init redis client, url=%s", config.Redis.Url)
	client := goredislib.NewClient(&goredislib.Options{
		Addr: config.Redis.Url,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logutil.GetLogger().Fatalf("init redis client error, err=%s", err)
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	logutil.GetLogger().Infof("init iot sdk, url=%s", config.Rabbitmq.Url)
	iot, err := iotsdk.New(config.Rabbitmq.Url, "web", client)
	if err != nil {
		logutil.GetLogger().Fatalf("init iot sdk error, err=%s", err)
	}

	tokenMaker, err := token.NewJWTMaker(config.Token.SymmetricKey)
	if err != nil {
		logutil.GetLogger().Fatalf("new jwt maker error, err=%s", err)
	}

	server, err := web.New(config, store, rs, tokenMaker, iot)
	if err != nil {
		logutil.GetLogger().Fatalf("init http server error, err=%s", err)
	}

	go func() {
		logutil.GetLogger().Infof("start http server, port=%s", config.Web.Port)
		err = server.Start("0.0.0.0:" + config.Web.Port)
		if err != nil {
			logutil.GetLogger().Fatalf("start http server error, err=%s", err)
		}
	}()
	defer func() {
		logutil.GetLogger().Info("shut down http server")
		if err := server.Shutdown(); err != nil {
			logutil.GetLogger().Errorf("shut down http server error, err=%s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
