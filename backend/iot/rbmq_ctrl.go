package iot

import (
	logutil "backend/util/log"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RbmqCtrl struct {
	edgeMapService *EdgeMapService
	resRepo        *RbmqRepo
	handlers       map[string]func(RbmqRequest) (any, *Error, bool)
}

func NewRbmqCtrl(ems *EdgeMapService, r *RbmqRepo) *RbmqCtrl {
	c := &RbmqCtrl{edgeMapService: ems, resRepo: r}
	c.handlers = map[string]func(RbmqRequest) (any, *Error, bool){
		"get-edge-system-info":        c.getEdgeSystemInfo,
		"add-points-to-coin-acceptor": c.addPointsToCoinAcceptor,
		"get-coin-acceptor-info":      c.getCoinAcceptorInfo,
		"get-coin-acceptor-status":    c.getCoinAcceptorStatus,
		"blink-coin-acceptor":         c.BlinkCoinAcceptor,
	}
	return c
}

func (c *RbmqCtrl) HandleRequest(bytes []byte) {
	var m1 MessageType1[RbmqRequest]
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

	var m2 MessageType2[any]
	m2.Type, m2.CorrID = m1.Type, m1.CorrID
	m2.Ts2 = time.Now().UnixMilli()

	res, e, ignored := handler(m1.Request)
	if ignored {
		return
	}
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

func (c *RbmqCtrl) getEdgeSystemInfo(req RbmqRequest) (any, *Error, bool) {
	if req.StoreID == nil || *req.StoreID == "" {
		return nil, nil, true
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		return nil, nil, true
	}

	edge := c.edgeMapService.Get(storeID)
	if edge == nil {
		return nil, nil, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	info, err := edge.GetEdgeSystemInfo(ctx)
	if err != nil {
		if err == ErrRPCRequestTimeout {
			return nil, &Error{Code: codeInternalError, Message: "internal error"}, false
		}
		logutil.GetLogger().Errorf("get edge system info error, err=%s, store_id=%s", err, storeID)
		return nil, &Error{Code: codeInternalError, Message: fmt.Sprintf("get edge system info error, store_id=%s", storeID)}, false
	}

	return map[string]any{
		"edge_version": info.EdgeVersion,
	}, nil, false
}

func (c *RbmqCtrl) addPointsToCoinAcceptor(req RbmqRequest) (any, *Error, bool) {
	if req.StoreID == nil || *req.StoreID == "" {
		return nil, nil, true
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		return nil, nil, true
	}

	edge := c.edgeMapService.Get(storeID)
	if edge == nil {
		return nil, nil, true
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}, false
	}

	if req.Amount == nil || *req.Amount <= 0 {
		return nil, &Error{Code: codeInvalidParameterError, Message: "amount is null or smaller than or equal to 0"}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := edge.AddPointsToCoinAcceptor(ctx, *req.DeviceID, *req.Amount); err != nil {
		if err == ErrRPCRequestTimeout {
			return nil, &Error{Code: codeInternalError, Message: "internal error"}, false
		}
		switch err.(type) {
		case *DeviceNotFoundError:
			return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device not found, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		default:
			logutil.GetLogger().Errorf("add points to coin acceptor error, err=%s, store_id=%s, device_id=%s", err, storeID, *req.DeviceID)
			return nil, &Error{Code: codeInternalError, Message: fmt.Sprintf("add points to coin acceptor error, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		}
	}

	return struct{}{}, nil, false
}

func (c *RbmqCtrl) getCoinAcceptorInfo(req RbmqRequest) (any, *Error, bool) {
	if req.StoreID == nil || *req.StoreID == "" {
		return nil, nil, true
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		return nil, nil, true
	}

	edge := c.edgeMapService.Get(storeID)
	if edge == nil {
		return nil, nil, true
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	info, err := edge.GetCoinAcceptorInfo(ctx, *req.DeviceID)
	if err != nil {
		if err == ErrRPCRequestTimeout {
			return nil, &Error{Code: codeInternalError, Message: "internal error"}, false
		}
		switch err.(type) {
		case *DeviceNotFoundError:
			return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device not found, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		default:
			logutil.GetLogger().Errorf("get coin acceptor info error, err=%s, store_id=%s, device_id=%s", err, storeID, *req.DeviceID)
			return nil, &Error{Code: codeInternalError, Message: fmt.Sprintf("get coin acceptor info error, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		}
	}

	return info, nil, false
}

func (c *RbmqCtrl) getCoinAcceptorStatus(req RbmqRequest) (any, *Error, bool) {
	if req.StoreID == nil || *req.StoreID == "" {
		return nil, nil, true
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		return nil, nil, true
	}

	edge := c.edgeMapService.Get(storeID)
	if edge == nil {
		return nil, nil, true
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := edge.GetCoinAcceptorStatus(ctx, *req.DeviceID)
	if err != nil {
		if err == ErrRPCRequestTimeout {
			return nil, &Error{Code: codeInternalError, Message: "internal error"}, false
		}
		switch err.(type) {
		case *DeviceNotFoundError:
			return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device not found, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		default:
			logutil.GetLogger().Errorf("get coin acceptor status error, err=%s, store_id=%s, device_id=%s", err, storeID, *req.DeviceID)
			return nil, &Error{Code: codeInternalError, Message: fmt.Sprintf("get coin acceptor status error, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		}
	}

	return status, nil, false
}

func (c *RbmqCtrl) BlinkCoinAcceptor(req RbmqRequest) (any, *Error, bool) {
	if req.StoreID == nil || *req.StoreID == "" {
		return nil, nil, true
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		return nil, nil, true
	}

	edge := c.edgeMapService.Get(storeID)
	if edge == nil {
		return nil, nil, true
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		return nil, &Error{Code: codeInvalidParameterError, Message: "device_id is null or empty"}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := edge.BlinkCoinAcceptor(ctx, *req.DeviceID); err != nil {
		if err == ErrRPCRequestTimeout {
			return nil, &Error{Code: codeInternalError, Message: "internal error"}, false
		}
		switch err.(type) {
		case *DeviceNotFoundError:
			return nil, &Error{Code: codeDeviceNotFoundError, Message: fmt.Sprintf("device not found, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		default:
			logutil.GetLogger().Errorf("blink coin acceptor error, err=%s, store_id=%s, device_id=%s", err, storeID, *req.DeviceID)
			return nil, &Error{Code: codeInternalError, Message: fmt.Sprintf("blink coin acceptor error, store_id=%s, device_id=%s", storeID, *req.DeviceID)}, false
		}
	}

	return struct{}{}, nil, false
}
