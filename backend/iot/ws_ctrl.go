package iot

import (
	db "backend/db/sqlc"
	"backend/heartbeat"
	distlockutil "backend/util/distlock"
	fsmutil "backend/util/fsm"
	logutil "backend/util/log"
	passwordutil "backend/util/password"
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	goredislib "github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// http to ws ctrl
type WsCtrl struct {
	store          db.IStore
	redisClient    *goredislib.Client
	redisSync      *redsync.Redsync
	edgeMapService *EdgeMapService
	storeEventRepo *RbmqRepo
}

func NewWsCtrl(store db.IStore, rc *goredislib.Client, rs *redsync.Redsync, ems *EdgeMapService, r *RbmqRepo) *WsCtrl {
	return &WsCtrl{store: store, redisClient: rc, redisSync: rs, edgeMapService: ems, storeEventRepo: r}
}

func (ctrl *WsCtrl) HandleRequest(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var storeID uuid.UUID
	loggedInChan := make(chan bool, 1)
	m2 := MessageType2[any]{}
	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logutil.GetLogger().Infof("websocket unexpected close, err=%s", err)
			}
			return
		}
		m1 := MessageType1[WsRequest]{}
		if err := json.Unmarshal(msg, &m1); err != nil {
			logutil.GetLogger().Errorf("json unmarshal error, msg=%#v, err=%s", msg, err)
			loggedInChan <- false
			return
		}
		// TODO: log request
		m2.Ts2 = time.Now().UnixMilli()

		if m1.Type != "login" {
			loggedInChan <- false
			return
		}
		m2.Type = m1.Type

		if m1.CorrID == "" {
			loggedInChan <- false
			return
		}
		m2.CorrID = m1.CorrID

		if m1.Request.StoreID == nil || *m1.Request.StoreID == "" {
			m2.Error = &Error{Code: codeInvalidParameterError, Message: "store_id is null or empty"}
			loggedInChan <- false
			return
		}
		if m1.Request.Password == nil || *m1.Request.Password == "" {
			m2.Error = &Error{Code: codeInvalidParameterError, Message: "password is null or empty"}
			loggedInChan <- false
			return
		}

		storeID, err = uuid.Parse(*m1.Request.StoreID)
		if err != nil {
			m2.Error = &Error{Code: codeStoreNotFoundError, Message: "store is not found"}
			loggedInChan <- false
			return
		}

		store, err := ctrl.store.GetStore(c, storeID)
		if err != nil {
			if err == sql.ErrNoRows {
				m2.Error = &Error{Code: codeStoreNotFoundError, Message: "store is not found"}
				loggedInChan <- false
				return
			}
			logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID)
			m2.Error = &Error{Code: codeInternalError, Message: "internal error"}
			loggedInChan <- false
			return
		}

		if !store.Password.Valid || passwordutil.CheckPassword(*m1.Request.Password, store.Password.String) != nil {
			m2.Error = &Error{Code: codeInvalidParameterError, Message: "wrong password"}
			loggedInChan <- false
			return
		}
		m2.Response = struct{}{}
		loggedInChan <- true
	}()

	select {
	case <-time.After(2 * time.Second):
		return
	case loggedIn := <-loggedInChan:
		m2.Ts3 = time.Now().UnixMilli()
		j, _ := json.Marshal(m2)
		conn.WriteMessage(websocket.TextMessage, j)
		if !loggedIn {
			return
		}
		// TODO: log response
	}

	m := ctrl.redisSync.NewMutex(distlockutil.GetEdgeStoreIDMutexName(storeID.String()))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetEdgeStoreIDMutexName(storeID.String()))
		return
	}

	unlock := func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetEdgeStoreIDMutexName(storeID.String()))
			return
		}
	}

	exist, err := heartbeat.CheckHeartbeat(ctrl.redisClient, heartbeat.GetStoreIDHeartbeatName(storeID.String()))
	if err != nil {
		logutil.GetLogger().Errorf("check heartbeat error, err=%s, heartbeat_name=%s", err, heartbeat.GetStoreIDHeartbeatName(storeID.String()))
		unlock()
		return
	}
	if exist {
		logutil.GetLogger().Warnf("heartbeat exist, heartbeat_name=%s", heartbeat.GetStoreIDHeartbeatName(storeID.String()))
		unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logutil.GetLogger().Infof("start heartbeat, heartbeat_name=%s", heartbeat.GetStoreIDHeartbeatName(storeID.String()))
		ticker := time.NewTicker(2900 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				if err := heartbeat.SendHeartbeat(ctrl.redisClient, heartbeat.GetStoreIDHeartbeatName(storeID.String()), 3*time.Second); err != nil {
					logutil.GetLogger().Errorf("send heartbeat error, err=%s, heartbeat_name=%s", err, heartbeat.GetStoreIDHeartbeatName(storeID.String()))
				}
			case <-ctx.Done():
				ticker.Stop()
				if err := heartbeat.StopHeartbeat(ctrl.redisClient, heartbeat.GetStoreIDHeartbeatName(storeID.String())); err != nil {
					logutil.GetLogger().Errorf("stop heartbeat error, err=%s, heartbeat_name=%s", err, heartbeat.GetStoreIDHeartbeatName(storeID.String()))
				}
				logutil.GetLogger().Infof("stop send heartbeat loop, heartbeat_name=%s", heartbeat.GetStoreIDHeartbeatName(storeID.String()))
				return
			}
		}
	}()

	unlock()

	edge := NewEdge(ctrl.store, storeID, c.Request.UserAgent(), c.ClientIP(), conn, ctrl.storeEventRepo)

	ctrl.edgeMapService.Add(storeID, edge)
	defer ctrl.edgeMapService.Delete(storeID)

	closedChan := make(chan struct{})
	go edge.readLoop(closedChan)
	edge.writeLoop(closedChan)
}

// ws server ctrl
type iotWsCtrl struct {
	store     db.IStore
	storeID   uuid.UUID
	userAgent string
	clientIp  string

	handlers     map[string]func(req WsRequest) (any, *Error)
	toClientChan chan []byte
}

func newIotWsCtrl(store db.IStore, storeID uuid.UUID, userAgent, clientIp string, toClientChan chan []byte) *iotWsCtrl {
	c := iotWsCtrl{
		store:        store,
		storeID:      storeID,
		userAgent:    userAgent,
		clientIp:     clientIp,
		toClientChan: toClientChan,
	}
	c.handlers = map[string]func(req WsRequest) (any, *Error){
		"register-coin-acceptor":                 c.registerCoinAcceptor,
		"add-coin-acceptor-coin-inserted-record": c.addCoinAcceptorCoinInsertedRecord,
	}
	return &c
}

func (c *iotWsCtrl) handleRequest(bytes []byte, m1 MessageType1[WsRequest]) {
	if m1.CorrID == "" {
		logutil.GetLogger().Errorf("corr id is empty, bytes=%s", string(bytes))
		return
	}
	handler, ok := c.handlers[m1.Type]
	if !ok {
		logutil.GetLogger().Errorf("unknown type, bytes=%s", string(bytes))
		return
	}
	// TODO: log request

	var m2 MessageType2[any]
	m2.Type, m2.CorrID = m1.Type, m1.CorrID
	m2.Ts2 = time.Now().UnixMilli()

	res, e := handler(m1.Request)
	m2.Ts3 = time.Now().UnixMilli()
	if e == nil {
		m2.Response = res
	} else {
		m2.Error = e
	}

	j, err := json.Marshal(m2)
	if err != nil {
		logutil.GetLogger().Errorf("json marshal error, m2=%#v, err=%s", m2, err)
		return
	}

	c.toClientChan <- j
	// TODO: log response
}

func (c *iotWsCtrl) registerCoinAcceptor(req WsRequest) (any, *Error) {
	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}
	}

	arg := db.CreateStoreDeviceWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       db.StoreDeviceChangedTypeCreate,
		ChangedBy:        uuid.NullUUID{Valid: false},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.userAgent},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.clientIp},
		StoreID:          c.storeID,
		DeviceID:         *req.DeviceID,
		Name:             "洗衣機",
		RealType:         db.StoreDeviceRealTypeCoinAcceptor,
		DisplayType:      db.StoreDeviceDisplayTypeWasher,
		State:            fsmutil.StoreDeviceStateActive,
	}

	if _, err := c.store.CreateStoreDeviceWithLog(context.Background(), arg); err != nil {
		if err == sql.ErrNoRows {
			return struct{}{}, nil
		}
		logutil.GetLogger().Errorf("create store device with log error, err=%s, arg=%#v", err, arg)
		return nil, &Error{Code: codeInternalError, Message: "internal error"}
	}
	return struct{}{}, nil
}

func (c *iotWsCtrl) addCoinAcceptorCoinInsertedRecord(req WsRequest) (any, *Error) {
	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}
	}

	if req.RecordID == nil || *req.RecordID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "record_id is null or empty"}
	}

	if req.Amount == nil || *req.Amount <= 0 {
		return nil, &Error{Code: codeInvalidParameterError, Message: "amount is null or smaller than or equal to 0"}
	}

	if req.Ts == nil || *req.Ts <= 0 {
		return nil, &Error{Code: codeInvalidParameterError, Message: "ts is null or smaller than or equal to 0"}
	}

	arg := db.CreateRecordParams{
		CreatedUserAgent: sql.NullString{Valid: true, String: c.userAgent},
		CreatedClientIp:  sql.NullString{Valid: true, String: c.clientIp},
		Type:             db.RecordTypeCoinAcceptorCoinInserted,
		StoreID:          c.storeID,
		RecordID:         sql.NullString{Valid: true, String: *req.RecordID},
		DeviceID:         sql.NullString{Valid: true, String: *req.DeviceID},
		Amount:           *req.Amount,
		Ts:               *req.Ts,
	}

	if _, err := c.store.CreateRecord(context.Background(), arg); err != nil {
		if err == sql.ErrNoRows {
			return struct{}{}, nil
		}
		logutil.GetLogger().Errorf("create record error, err=%s, arg=%#v", err, arg)
		return nil, &Error{Code: codeInternalError, Message: "internal error"}
	}
	return struct{}{}, nil
}
