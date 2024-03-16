package main

import (
	db "backend/db/sqlc"
	"backend/iot"
	"backend/server"
	configutil "backend/util/config"
	logutil "backend/util/log"
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	_ "github.com/jackc/pgx/v4/stdlib"
	goredislib "github.com/redis/go-redis/v9"
)

var version string

func main() {
	logutil.GetLogger().Info("iot-backend, version: ", version)

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
	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr: config.Redis.Url,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logutil.GetLogger().Fatalf("init redis client error, err=%s", err)
	}

	pool := goredis.NewPool(redisClient)
	redisSync := redsync.New(pool)

	edgeMapService := iot.NewEdgeMapService()

	logutil.GetLogger().Infof("new rbmq store event repo, url=%s, exchange=%s, key=%s", config.Rabbitmq.Url, iot.Exchange, iot.ResponseKeyFmt)
	storeEventRepo, err := iot.NewRbmqRepo(config.Rabbitmq.Url, iot.Exchange, iot.StoreEventKeyFmt)
	if err != nil {
		logutil.GetLogger().Fatalf("new rbmq store event repo error, err=%s", err)
	}
	defer func() {
		logutil.GetLogger().Infof("close rbmq store event repo")
		storeEventRepo.Close()
	}()

	wsCtrl := iot.NewWsCtrl(store, redisClient, redisSync, edgeMapService, storeEventRepo)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/v1/ws", wsCtrl.HandleRequest)

	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + config.Iot.Port,
		Handler: router,
	}
	go func() {
		logutil.GetLogger().Infof("start http server, port=%s", config.Iot.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logutil.GetLogger().Fatalf("start http server error, err=%s", err)
		}
	}()
	defer func() {
		logutil.GetLogger().Info("shut down http server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			logutil.GetLogger().Errorf("shut down http server error, err=%s", err)
		}
	}()

	logutil.GetLogger().Infof("new rbmq response repo, url=%s, exchange=%s, key=%s", config.Rabbitmq.Url, iot.Exchange, iot.ResponseKeyFmt)
	resRepo, err := iot.NewRbmqRepo(config.Rabbitmq.Url, iot.Exchange, iot.ResponseKeyFmt)
	if err != nil {
		logutil.GetLogger().Fatalf("new rbmq response repo error, err=%s", err)
	}
	defer func() {
		logutil.GetLogger().Infof("close rbmq response repo")
		resRepo.Close()
	}()

	rbmqCtrl := iot.NewRbmqCtrl(edgeMapService, resRepo)

	rbmqServer := &server.RbmqServer{
		Url:      config.Rabbitmq.Url,
		Exchange: iot.Exchange,
		Key:      iot.RequestKey,
		Handler:  rbmqCtrl.HandleRequest,
	}

	go func() {
		logutil.GetLogger().Infof("start rbmq server, url=%s, exchange=%s, key=%s", rbmqServer.Url, rbmqServer.Exchange, rbmqServer.Key)
		if err := rbmqServer.ListenAndServe(); err != nil {
			logutil.GetLogger().Fatalf("run rbmq server error, err=%s", err)
		}
	}()
	defer func() {
		logutil.GetLogger().Infof("shut down rbmq server")
		rbmqServer.Shutdown()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
