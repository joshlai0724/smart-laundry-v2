package edge

import (
	"context"
	infoutil "edge/util/info"
	logutil "edge/util/log"
	"encoding/json"
	"fmt"
	"time"
)

type WsCtrl struct {
	systemInfo       infoutil.Info
	deviceMapService *DeviceMapService
	toClientChan     chan []byte
	handlers         map[string]func(req WsRequest) (any, *Error)
}

func NewWsCtrl(si infoutil.Info, dms *DeviceMapService, toClientChan chan []byte) *WsCtrl {
	c := WsCtrl{
		systemInfo:       si,
		deviceMapService: dms,
		toClientChan:     toClientChan,
	}
	c.handlers = map[string]func(req WsRequest) (any, *Error){
		"get-edge-system-info":        c.getEdgeSystemInfo,
		"add-points-to-coin-acceptor": c.addPointsToCoinAcceptor,
		"get-coin-acceptor-info":      c.getCoinAcceptorInfo,
		"get-coin-acceptor-status":    c.getCoinAcceptorStatus,
		"blink-coin-acceptor":         c.blinkCoinAcceptor,
	}
	return &c
}

func (c *WsCtrl) handleRequest(bytes []byte, m1 MessageType1[WsRequest]) {
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

func (c *WsCtrl) getEdgeSystemInfo(req WsRequest) (any, *Error) {
	return map[string]any{
		"edge_version": c.systemInfo.EdgeVersion,
	}, nil
}

func (c *WsCtrl) addPointsToCoinAcceptor(req WsRequest) (any, *Error) {
	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}
	}

	if req.Amount == nil {
		return nil, &Error{Code: codeInvalidParameterError, Message: "amount is null"}
	}

	coinAcceptor := c.deviceMapService.GetCoinAcceptor(*req.DeviceID)
	if coinAcceptor == nil {
		return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device is not found, device_id=%s", *req.DeviceID)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := coinAcceptor.AddPoints(ctx, *req.Amount); err != nil {
		switch err.(type) {
		case *InvalidParameterError:
			return nil, &Error{Code: codeInvalidParameterError, Message: err.Error()}
		case *IllegalStateError:
			return nil, &Error{Code: codeIllegalStateError, Message: err.Error()}
		default:
			logutil.GetLogger().Errorf("add points to coin acceptor error, err=%s, device_id=%s, amount=%d",
				err, *req.DeviceID, *req.Amount)
			return nil, &Error{Code: codeInternalError, Message: err.Error()}
		}
	}
	return struct{}{}, nil
}

func (c *WsCtrl) getCoinAcceptorInfo(req WsRequest) (any, *Error) {
	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}
	}

	coinAcceptor := c.deviceMapService.GetCoinAcceptor(*req.DeviceID)
	if coinAcceptor == nil {
		return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device is not found, device_id=%s", *req.DeviceID)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	info, err := coinAcceptor.GetDeviceInfo(ctx)
	if err != nil {
		logutil.GetLogger().Errorf("get coin acceptor info error, err=%s, device_id=%s", err, *req.DeviceID)
		return nil, &Error{Code: codeInternalError, Message: err.Error()}
	}
	return map[string]any{
		"firmware_version": info.FirmwareVersion,
	}, nil
}

func (c *WsCtrl) getCoinAcceptorStatus(req WsRequest) (any, *Error) {
	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}
	}

	coinAcceptor := c.deviceMapService.GetCoinAcceptor(*req.DeviceID)
	if coinAcceptor == nil {
		return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device is not found, device_id=%s", *req.DeviceID)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := coinAcceptor.GetDeviceStatus(ctx)
	if err != nil {
		logutil.GetLogger().Errorf("get coin acceptor status error, err=%s, device_id=%s", err, *req.DeviceID)
		return nil, &Error{Code: codeInternalError, Message: err.Error()}
	}
	return map[string]any{
		"points": status.Points,
		"state":  status.State,
		"ts":     status.Ts,
	}, nil
}

func (c *WsCtrl) blinkCoinAcceptor(req WsRequest) (any, *Error) {
	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}
	}

	coinAcceptor := c.deviceMapService.GetCoinAcceptor(*req.DeviceID)
	if coinAcceptor == nil {
		return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device is not found, device_id=%s", *req.DeviceID)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := coinAcceptor.Blink(ctx); err != nil {
		logutil.GetLogger().Errorf("blink coin acceptor error, err=%s, device_id=%s", err, *req.DeviceID)
		return nil, &Error{Code: codeInternalError, Message: err.Error()}
	}
	return struct{}{}, nil
}
