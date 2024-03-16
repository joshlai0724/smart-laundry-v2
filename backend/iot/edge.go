package iot

import (
	db "backend/db/sqlc"
	logutil "backend/util/log"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Edge struct {
	// storeID
	conn *websocket.Conn

	toClientChan chan []byte

	rpcRepo RpcRepo

	requestHandler  func(bytes []byte, m1 MessageType1[WsRequest])
	responseHandler func(bytes []byte, m2 MessageType2[WsResponse])
	eventHandler    func(bytes []byte, m3 MessageType3[WsEvent])
}

func NewEdge(store db.IStore, storeID uuid.UUID, userAgent, clientIp string, conn *websocket.Conn, r *RbmqRepo) *Edge {
	c := &Edge{conn: conn}
	c.toClientChan = make(chan []byte, 100)
	rpcRepo := newRpcRepo(c.toClientChan)
	c.rpcRepo = rpcRepo

	c.requestHandler = newIotWsCtrl(store, storeID, userAgent, clientIp, c.toClientChan).handleRequest
	c.responseHandler = rpcRepo.handleResponse
	c.eventHandler = newEdgeEventCtrl(r, storeID).handleEvent
	return c
}

func (c *Edge) readLoop(closedChan chan struct{}) {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			closedChan <- struct{}{}
			break
		}
		go c.handleMessage(msg)
	}
}

func (c *Edge) handleMessage(bytes []byte) {
	var m Message[WsRequest, WsResponse, WsEvent]
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
	} else if m.Event != nil {
		c.eventHandler(bytes, MessageType3[WsEvent]{
			Type:  m.Type,
			Event: *m.Event,
			Ts3:   m.Ts3,
		})
	} else {
		logutil.GetLogger().Errorf("unknown type, bytes=%s", string(bytes))
	}
}

func (c *Edge) writeLoop(closedChan chan struct{}) {
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

func (c *Edge) GetEdgeSystemInfo(ctx context.Context) (*EdgeSystemInfo, error) {
	m2, err := Rpc[struct{}](
		ctx,
		c.rpcRepo,
		"get-edge-system-info",
		struct{}{})

	if err != nil {
		return nil, err
	}

	return &EdgeSystemInfo{
		EdgeVersion: m2.Response.EdgeVersion,
	}, nil
}

type addPointsToCoinAcceptorRequest struct {
	DeviceID string `json:"device_id"`
	Amount   int32  `json:"amount"`
}

func (c *Edge) AddPointsToCoinAcceptor(ctx context.Context, deviceID string, amount int32) error {
	_, err := Rpc[addPointsToCoinAcceptorRequest](
		ctx,
		c.rpcRepo,
		"add-points-to-coin-acceptor",
		addPointsToCoinAcceptorRequest{
			DeviceID: deviceID,
			Amount:   amount,
		})
	return err
}

type getCoinAcceptorInfoRequest struct {
	DeviceID string `json:"device_id"`
}

func (c *Edge) GetCoinAcceptorInfo(ctx context.Context, deviceID string) (*CoinAcceptorInfo, error) {
	m2, err := Rpc[getCoinAcceptorInfoRequest](
		ctx,
		c.rpcRepo,
		"get-coin-acceptor-info",
		getCoinAcceptorInfoRequest{
			DeviceID: deviceID,
		})

	if err != nil {
		return nil, err
	}

	return &CoinAcceptorInfo{
		FirmwareVersion: m2.Response.FirmwareVersion,
	}, nil
}

type getCoinAcceptorStatusRequest struct {
	DeviceID string `json:"device_id"`
}

func (c *Edge) GetCoinAcceptorStatus(ctx context.Context, deviceID string) (*CoinAcceptorStatus, error) {
	m2, err := Rpc[getCoinAcceptorStatusRequest](
		ctx,
		c.rpcRepo,
		"get-coin-acceptor-status",
		getCoinAcceptorStatusRequest{
			DeviceID: deviceID,
		})

	if err != nil {
		return nil, err
	}

	return &CoinAcceptorStatus{
		Points: m2.Response.Points,
		State:  m2.Response.State,
		Ts:     m2.Response.Ts,
	}, nil
}

type blinkCoinAcceptorRequest struct {
	DeviceID string `json:"device_id"`
}

func (c *Edge) BlinkCoinAcceptor(ctx context.Context, deviceID string) error {
	_, err := Rpc[blinkCoinAcceptorRequest](
		ctx,
		c.rpcRepo,
		"blink-coin-acceptor",
		blinkCoinAcceptorRequest{
			DeviceID: deviceID,
		})

	if err != nil {
		return err
	}

	return nil
}
