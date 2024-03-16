package simulator

import (
	logutil "coin-acceptor/util/log"
	"encoding/json"
	"math/rand"
	"time"
)

type CoinAccpetorCtrl struct {
	resRepo      *CoinAcceptorRepo
	coinAcceptor *CoinAcceptor
	handlers     map[string]func(Request) (any, *Error)
}

func NewCoinAcceptorCtrl(r *CoinAcceptorRepo, ca *CoinAcceptor) *CoinAccpetorCtrl {
	c := &CoinAccpetorCtrl{resRepo: r, coinAcceptor: ca}
	c.handlers = map[string]func(Request) (any, *Error){
		"add-points":        c.addPoints,
		"get-device-info":   c.getDeviceInfo,
		"get-device-status": c.getDeviceStatus,
		"get-sensor-values": c.getSensorValues,
		"check-health":      c.checkHealth,
		"blink":             c.blink,
	}
	return c
}

func (c *CoinAccpetorCtrl) HandleRequest(bytes []byte) {
	var m1 MessageType1
	if err := json.Unmarshal(bytes, &m1); err != nil {
		logutil.GetLogger().Errorf("json unmarshal error, bytes=%#v, err=%s", bytes, err)
		return
	}

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

	var m2 MessageType2
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

	c.resRepo.Publish(m2.CorrID, j)
	// TODO: log response
}

func (c *CoinAccpetorCtrl) addPoints(req Request) (any, *Error) {
	if req.Amount == nil {
		return nil, &Error{Code: codeInvalidParameterError, Message: "amount is null"}
	}
	if *req.Amount <= 0 {
		return nil, &Error{Code: codeInvalidParameterError, Message: "amount is smaller than or equal to 0"}
	}
	c.coinAcceptor.AddPoints(*req.Amount)
	return struct{}{}, nil
}

func (c *CoinAccpetorCtrl) getDeviceInfo(req Request) (any, *Error) {
	info := c.coinAcceptor.GetDeviceInfo()
	return map[string]any{
		"firmware_version": info.FirmwareVersion,
	}, nil
}

func (c *CoinAccpetorCtrl) getDeviceStatus(req Request) (any, *Error) {
	status, ts := c.coinAcceptor.GetDeviceStatus()
	return map[string]any{
		"points": status.Points,
		"state":  status.State,
		"ts":     ts,
	}, nil
}

func (c *CoinAccpetorCtrl) getSensorValues(req Request) (any, *Error) {
	return map[string]any{
		"wifi_signal_strength": rand.Float64()*60 - 60,
	}, nil
}

func (c *CoinAccpetorCtrl) checkHealth(req Request) (any, *Error) {
	return struct{}{}, nil
}

func (c *CoinAccpetorCtrl) blink(req Request) (any, *Error) {
	logutil.GetLogger().Infof("I am blinking!")
	return struct{}{}, nil
}
