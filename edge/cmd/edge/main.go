package main

import (
	"context"
	"database/sql"
	db "edge/db/sqlc"
	"edge/edge"
	configutil "edge/util/config"
	infoutil "edge/util/info"
	logutil "edge/util/log"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var version string

func main() {
	logutil.GetLogger().Info("edge-backend, version: ", version)

	config, err := configutil.Load(os.Getenv("ENV"))
	if err != nil {
		logutil.GetLogger().Fatalf("load config error, err=%s", err)
	}
	logutil.GetLogger().Infof("configFile=%s", configutil.GetConfigFile())

	systemInfo := infoutil.ReadInfo(config.Info.EdgeVersionFile)

	logutil.GetLogger().Infof("init db connection, source=%s", config.DB.Source)
	conn, err := sql.Open("pgx", config.DB.Source)
	if err != nil {
		logutil.GetLogger().Fatalf("init db connection error, err=%s", err)
	}
	if err := conn.Ping(); err != nil {
		logutil.GetLogger().Fatalf("init db connection error, err=%s", err)
	}

	store := db.NewStore(conn)

	deviceMapService := edge.NewDeviceMapService()

	iotContainer := &edge.IotContainer{}

	go func() {
		for {
			// TODO: graceful shut down
			func() {
				conn, _, err := websocket.DefaultDialer.Dial(config.Iot.Url, nil)
				if err != nil {
					logutil.GetLogger().Warnf("connect to iot backend err, err=%s, url=%s", err, config.Iot.Url)
					return
				}
				defer conn.Close()
				logutil.GetLogger().Infof("connect to iot backend, url=%s", config.Iot.Url)

				iot := edge.NewIot(systemInfo, deviceMapService, conn)

				done := make(chan struct{}, 1)
				go func() {
					closedChan := make(chan struct{})
					go iot.ReadLoop(closedChan)
					iot.WriteLoop(closedChan)
					logutil.GetLogger().Infof("disconnect to iot backend")
					done <- struct{}{}
				}()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				if err := iot.Login(ctx, config.StoreID, config.Password); err != nil {
					logutil.GetLogger().Warnf("login to iot backend error, err=%s, store_id=%s", err, config.StoreID)
					return
				}
				logutil.GetLogger().Infof("login to iot backend success, store_id=%s", config.StoreID)

				iotContainer.Set(iot)
				defer iotContainer.Clear()

				<-done
			}()
			time.Sleep(config.Iot.RetryInterval)
		}
	}()

	serviceDiscovery := edge.NewServiceDiscoveryCtrl(deviceMapService)
	coinAcceptorEventCtrl := edge.NewCoinAcceptorEvenCtrl(store, iotContainer)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Mosquitto.Url)
	opts.SetClientID(config.StoreID)
	opts.SetUsername(config.Mosquitto.Username)
	opts.SetPassword(config.Mosquitto.Password)
	opts.OnConnect = func(client mqtt.Client) {
		logutil.GetLogger().Infof("connect to mosquitto, url=%s, client_id=%s", config.Mosquitto.Url, config.StoreID)
		client.Subscribe(edge.CoinAcceptorEventKey, 2, func(c mqtt.Client, msg mqtt.Message) {
			m3 := edge.MessageType3[edge.MqttEvent]{}
			err := json.Unmarshal(msg.Payload(), &m3)
			if err != nil {
				return
			}
			switch m3.Type {
			case "beacon":
				go serviceDiscovery.HandleCoinAcceptorBeacon(c, m3.Event.DeviceID)
			case "coin-inserted":
				go coinAcceptorEventCtrl.HandleCoinInserted(m3.Event.DeviceID, m3.Event.Amount, m3.Event.Ts)
			case "device-status-changed":
				go coinAcceptorEventCtrl.HandleDeviceStatusChanged(m3.Event.DeviceID, edge.CoinAcceptorStatus{
					Points: m3.Event.Points, State: m3.Event.State, Ts: m3.Event.Ts})
			default:
				return
			}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go checkAndRegisterCoinAcceptors(ctx, deviceMapService, iotContainer)
	go uploadRecords(ctx, config, iotContainer, store)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func checkAndRegisterCoinAcceptors(ctx context.Context, s *edge.DeviceMapService, ic *edge.IotContainer) {
	map1 := map[string]bool{} // busy or not
	var m sync.Mutex

	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			for _, ca := range s.GetCoinAcceptorList() {
				m.Lock()
				if map1[ca.GetDeviceID()] {
					m.Unlock()
					continue
				}
				map1[ca.GetDeviceID()] = true
				go func(ca *edge.CoinAcceptor) {
					defer func() {
						m.Lock()
						map1[ca.GetDeviceID()] = false
						m.Unlock()
					}()
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					if err := ca.CheckHealth(ctx); err != nil {
						s.DeleteCoinAcceptor(ca.GetDeviceID())
						logutil.GetLogger().Infof("fail to check health and delete the coin acceptor, err=%s, device_id=%s", err, ca.GetDeviceID())
						return
					}

					iot := ic.Get()
					if iot == nil {
						return
					}

					if ca.Registered() {
						return
					}

					ctx, cancel = context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					if err := iot.RegisterCoinAcceptor(ctx, ca.GetDeviceID()); err != nil {
						logutil.GetLogger().Errorf("register coin acceptor error, err=%s, device_id=%s", err, ca.GetDeviceID())
						return
					}
					ca.SetRegistered(true)
				}(ca)
				m.Unlock()
			}
		}
	}
}

func uploadRecords(ctx context.Context, config configutil.Config, ic *edge.IotContainer, store db.IStore) {
	timer := time.NewTicker(config.RecordsResend.Interval)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			func() {
				iot := ic.Get()
				if iot == nil {
					return
				}

				records, err := store.GetUnuploadedRecords(context.Background(), config.RecordsResend.BatchSize)
				if err != nil {
					logutil.GetLogger().Errorf("get unuploaded records error, err=%s, limit=%d", err, config.RecordsResend.BatchSize)
					return
				}

				for _, record := range records {
					switch record.Type {
					case db.RecordTypeCoinAcceptorCoinInserted:
						ctx, cancel := context.WithTimeout(context.Background(), time.Second)
						defer cancel()

						if err := iot.AddCoinAcceptorCoinInsertedRecord(ctx, record.ID, record.DeviceID, record.Amount, record.Ts); err != nil {
							logutil.GetLogger().Errorf("upload coin acceptor coin inserted record error, err=%s, record_id=%s, device_id=%s, amount=%d, ts=%d",
								err, record.ID, record.DeviceID, record.Amount, record.Ts)
							return
						}

						arg := db.SetRecordIsUploadedParams{
							ID:         record.ID,
							IsUploaded: true,
							UploadedAt: sql.NullInt64{Valid: true, Int64: time.Now().UnixMilli()},
						}

						if err := store.SetRecordIsUploaded(context.Background(), arg); err != nil {
							logutil.GetLogger().Errorf("set record is uploaded error, err=%s, arg=%#v",
								err, arg)
							return
						}
					}

					logutil.GetLogger().Infof("reupload record success, record_id=%s", record.ID)
				}
			}()
		}
	}
}
