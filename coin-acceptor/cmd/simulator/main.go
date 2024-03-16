package main

import (
	"coin-acceptor/simulator"
	configutil "coin-acceptor/util/config"
	logutil "coin-acceptor/util/log"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var version string

func main() {
	logutil.GetLogger().Info("coin-acceptor-simulator, version: ", version)

	config, err := configutil.Load(os.Getenv("ENV"))
	if err != nil {
		logutil.GetLogger().Fatalf("load config error, err=%s", err)
	}
	logutil.GetLogger().Infof("configFile=%s", configutil.GetConfigFile())

	coinAcceptor := simulator.NewCoinAcceptor(config.DeviceID,
		simulator.CoinAcceptorInfo{FirmwareVersion: "v1.10.3"},
		simulator.CoinAcceptorStatus{
			Points: config.Points,
			State:  config.State,
		})

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Mosquitto.Url)
	opts.SetClientID(config.DeviceID)
	opts.SetUsername(config.Mosquitto.Username)
	opts.SetPassword(config.Mosquitto.Password)
	opts.OnConnect = func(client mqtt.Client) {
		logutil.GetLogger().Infof("connect to mosquitto, url=%s, client_id=%s", config.Mosquitto.Url, config.DeviceID)

		resRepo := simulator.NewCoinAcceptorRepo(
			client, strings.Replace(simulator.ResponseKeyFmt, "%s", config.DeviceID, 1))

		coinAcceptorCtrl := simulator.NewCoinAcceptorCtrl(resRepo, coinAcceptor)

		client.Subscribe(fmt.Sprintf(simulator.RequestKeyFmt, config.DeviceID), 2, func(c mqtt.Client, msg mqtt.Message) {
			coinAcceptorCtrl.HandleRequest(msg.Payload())
		})
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		logutil.GetLogger().Warnf("mosquitto connection lost, err=%s", err)
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logutil.GetLogger().Fatalf("connect to mosquitto error, err=%s, url=%s", token.Error(), config.Mosquitto.Url)
	}
	defer func() {
		logutil.GetLogger().Infof("disconnect to mosquitto")
		client.Disconnect(0)
	}()

	eventRepo := simulator.NewCoinAcceptorRepo(client, simulator.EventKey)
	coinAcceptor.SetEventRepo(eventRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logutil.GetLogger().Infof("start beacon, interval=%s", config.BeaconInterval)
		timer := time.NewTicker(config.BeaconInterval)
		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				m3 := simulator.MessageType3{
					Type: "beacon",
					Event: struct {
						DeviceID string `json:"device_id"`
					}{DeviceID: config.DeviceID},
					Ts3: time.Now().UnixMilli(),
				}
				j, _ := json.Marshal(m3)
				eventRepo.Publish(config.DeviceID, j)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
