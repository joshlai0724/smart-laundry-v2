package edge

import (
	"context"
	infoutil "edge/util/info"
	logutil "edge/util/log"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Iot struct {
	// storeID
	conn *websocket.Conn

	toClientChan chan []byte

	rpcRepo WsRpcRepo

	requestHandler  func(bytes []byte, m1 MessageType1[WsRequest])
	responseHandler func(bytes []byte, m2 MessageType2[WsResponse])
}

func NewIot(si infoutil.Info, dms *DeviceMapService, conn *websocket.Conn) *Iot {
	c := &Iot{conn: conn}
	c.toClientChan = make(chan []byte, 100)
	rpcRepo := newWsRpcRepo(c.toClientChan)
	c.rpcRepo = rpcRepo

	c.requestHandler = NewWsCtrl(si, dms, c.toClientChan).handleRequest
	c.responseHandler = rpcRepo.handleResponse
	return c
}

func (c *Iot) ReadLoop(closedChan chan struct{}) {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			closedChan <- struct{}{}
			break
		}
		go c.handleMessage(msg)
	}
}

func (c *Iot) handleMessage(bytes []byte) {
	var m Message[WsRequest, WsResponse]
	if err := json.Unmarshal(bytes, &m); err != nil {
		logutil.GetLogger().Errorf("json unmarshal error, bytes=%#v, err=%s", bytes, err)
		return
	}

	if m.Request != nil {
		c.requestHandler(bytes, MessageType1[WsRequest]{
			Type:    m.Type,
			CorrID:  m.CorrID,
			Request: *m.Request,
			Ts1:     m.Ts1,
		})
	} else if m.Response != nil || m.Error != nil {
		req := WsResponse{}
		if m.Response != nil {
			req = *m.Response
		}
		c.responseHandler(bytes, MessageType2[WsResponse]{
			Type:     m.Type,
			CorrID:   m.CorrID,
			Response: req,
			Error:    m.Error,
			Ts2:      m.Ts2,
			Ts3:      m.Ts3,
		})
	} else {
		logutil.GetLogger().Errorf("unknown type, bytes=%s", string(bytes))
	}
}

func (c *Iot) WriteLoop(closedChan chan struct{}) {
	for {
		select {
		case msg := <-c.toClientChan:
			c.conn.WriteMessage(websocket.TextMessage, msg)
		case <-closedChan:
			// TODO: close
			return
		}
	}
}

type loginRequest struct {
	StoreID  string `json:"store_id"`
	Password string `json:"password"`
}

func (c *Iot) Login(ctx context.Context, storeID string, password string) error {
	_, err := WsRpc[loginRequest](
		ctx,
		c.rpcRepo,
		"login",
		loginRequest{
			StoreID:  storeID,
			Password: password,
		})
	return err
}

type registerCoinAcceptorRequest struct {
	DeviceID string `json:"device_id"`
}

func (c *Iot) RegisterCoinAcceptor(ctx context.Context, deviceID string) error {
	_, err := WsRpc[registerCoinAcceptorRequest](
		ctx,
		c.rpcRepo,
		"register-coin-acceptor",
		registerCoinAcceptorRequest{
			DeviceID: deviceID,
		})
	return err
}

type addCoinAcceptorCoinInsertedRecordRequest struct {
	RecordID uuid.UUID `json:"record_id"`
	DeviceID string    `json:"device_id"`
	Amount   int32     `json:"amount"`
	Ts       int64     `json:"ts"`
}

func (c *Iot) AddCoinAcceptorCoinInsertedRecord(ctx context.Context, recordID uuid.UUID, deviceID string, amount int32, ts int64) error {
	_, err := WsRpc[addCoinAcceptorCoinInsertedRecordRequest](
		ctx,
		c.rpcRepo,
		"add-coin-acceptor-coin-inserted-record",
		addCoinAcceptorCoinInsertedRecordRequest{
			RecordID: recordID,
			DeviceID: deviceID,
			Amount:   amount,
			Ts:       ts,
		})
	return err
}

func (c *Iot) SendCoinAcceptorStatusChangedEvent(deviceID string, status CoinAcceptorStatus) {
	m3 := MessageType3[any]{
		Type: "coin-acceptor-status-changed",
		Event: struct {
			DeviceID string `json:"device_id"`
			Points   int32  `json:"points"`
			State    string `json:"state"`
			Ts       int64  `json:"ts"`
		}{DeviceID: deviceID, Points: status.Points, State: status.State, Ts: status.Ts},
		Ts3: time.Now().UnixMilli(),
	}
	j, _ := json.Marshal(m3)
	c.toClientChan <- j
}
