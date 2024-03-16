package web

import (
	db "backend/db/sqlc"
	iotsdk "backend/iot-sdk"
	"backend/token"
	distlockutil "backend/util/distlock"
	logutil "backend/util/log"
	roleutil "backend/util/role"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type getStoreDevicesUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) getStoreDevices(c *gin.Context) {
	var req getStoreDevicesUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store not found, store_id=%s", *req.StoreID)))
		return
	}

	storeDevices, err := s.store.GetStoreDevices(c, storeID)
	if err != nil {
		logutil.GetLogger().Errorf("get store devices error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	devices := make([]gin.H, 0, len(storeDevices))
	for _, storeDevice := range storeDevices {
		devices = append(devices, gin.H{
			"id":           storeDevice.DeviceID,
			"name":         storeDevice.Name,
			"real_type":    storeDevice.RealType,
			"display_type": storeDevice.DisplayType,
		})
	}
	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

type getStoreCoinAcceptorInfoUri struct {
	StoreID  *string `uri:"store_id"`
	DeviceID *string `uri:"device_id"`
}

func (s *Server) getStoreCoinAcceptorInfo(c *gin.Context) {
	var req getStoreCoinAcceptorInfoUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "device_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
		return
	}

	arg := db.GetStoreDeviceParams{
		StoreID:  storeID,
		DeviceID: *req.DeviceID,
	}

	storeDevice, err := s.store.GetStoreDevice(c, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreDeviceNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
			return
		}
		logutil.GetLogger().Errorf("get store device error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":         storeDevice.Name,
		"real_type":    storeDevice.RealType,
		"display_type": storeDevice.DisplayType,
	})
}

type updateStoreCoinAcceptorInfoUri struct {
	StoreID  *string `uri:"store_id"`
	DeviceID *string `uri:"device_id"`
}

type updateStoreCoinAcceptorInfoRequest struct {
	Name        *string `json:"name"`
	DisplayType *string `json:"display_type"`
}

func (s *Server) updateStoreCoinAcceptorInfo(c *gin.Context) {
	var reqUri updateStoreCoinAcceptorInfoUri
	if err := c.ShouldBindUri(&reqUri); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqUri.StoreID == nil || *reqUri.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if reqUri.DeviceID == nil || *reqUri.DeviceID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "device_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*reqUri.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *reqUri.StoreID, *reqUri.DeviceID)))
		return
	}

	var reqJson updateStoreCoinAcceptorInfoRequest
	if err := c.ShouldBindJSON(&reqJson); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqJson.Name == nil || *reqJson.Name == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is null or empty"))
		return
	}

	if reqJson.DisplayType == nil || *reqJson.DisplayType == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "display_type is null or empty"))
		return
	}

	if *reqJson.DisplayType != db.StoreDeviceDisplayTypeWasher && *reqJson.DisplayType != db.StoreDeviceDisplayTypeDryer {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, fmt.Sprintf("display type is invalid, display_type=%s", *reqJson.DisplayType)))
		return
	}

	arg1 := db.GetStoreDeviceParams{
		StoreID:  storeID,
		DeviceID: *reqUri.DeviceID,
	}

	storeDevice, err := s.store.GetStoreDevice(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreDeviceNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *reqUri.StoreID, *reqUri.DeviceID)))
			return
		}
		logutil.GetLogger().Errorf("get store device error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg2 := db.SetStoreDeviceNameAndDisplayTypeWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       db.StoreDeviceChangedTypeUpdateInfo,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		DeviceID:         *reqUri.DeviceID,
		Name:             *reqJson.Name,
		DisplayType:      *reqJson.DisplayType,
	}

	if arg2.Name == storeDevice.Name && arg2.DisplayType == storeDevice.DisplayType {
		c.Status(http.StatusNoContent)
		return
	}

	if err := s.store.SetStoreDeviceNameAndDisplayTypeWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store device name and display type with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type getStoreCoinAcceptorStatusUri struct {
	StoreID  *string `uri:"store_id"`
	DeviceID *string `uri:"device_id"`
}

func (s *Server) getStoreCoinAcceptorStatus(c *gin.Context) {
	var req getStoreCoinAcceptorStatusUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "device_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
		return
	}

	arg := db.GetStoreDeviceParams{
		StoreID:  storeID,
		DeviceID: *req.DeviceID,
	}

	if _, err := s.store.GetStoreDevice(c, arg); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreDeviceNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
			return
		}
		logutil.GetLogger().Errorf("get store device error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	ctx, cancel := context.WithTimeout(c, time.Second)
	defer cancel()

	status, err := s.iot.GetCoinAcceptorStatus(ctx, storeID, *req.DeviceID)
	if err != nil {
		switch err.(type) {
		case *iotsdk.DeviceNotFoundError:
			c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreDeviceNotOnlineError, fmt.Sprintf("store device is not online, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
		case *iotsdk.StoreNotFoundError:
			c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreNotOnlineError, fmt.Sprintf("store is not online, store_id=%s", *req.StoreID)))
		default:
			logutil.GetLogger().Errorf("get coin acceptor status error, err=%s, store_id=%s, device_id=%s", err, storeID, *req.DeviceID)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"points": status.Points,
		"state":  status.State,
		"ts":     status.Ts,
	})
}

type blinkStoreCoinAcceptorUri struct {
	StoreID  *string `uri:"store_id"`
	DeviceID *string `uri:"device_id"`
}

func (s *Server) blinkStoreCoinAcceptor(c *gin.Context) {
	var req blinkStoreCoinAcceptorUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "device_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
		return
	}

	arg := db.GetStoreDeviceParams{
		StoreID:  storeID,
		DeviceID: *req.DeviceID,
	}

	if _, err := s.store.GetStoreDevice(c, arg); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreDeviceNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
			return
		}
		logutil.GetLogger().Errorf("get store device error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	ctx, cancel := context.WithTimeout(c, time.Second)
	defer cancel()

	if err := s.iot.BlinkCoinAcceptor(ctx, storeID, *req.DeviceID); err != nil {
		switch err.(type) {
		case *iotsdk.DeviceNotFoundError:
			c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreDeviceNotOnlineError, fmt.Sprintf("store device is not online, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
		case *iotsdk.StoreNotFoundError:
			c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreNotOnlineError, fmt.Sprintf("store is not online, store_id=%s", *req.StoreID)))
		default:
			logutil.GetLogger().Errorf("blink coin acceptor error, err=%s, store_id=%s, device_id=%s", err, storeID, *req.DeviceID)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

type insertCoinToStoreCoinAcceptorUri struct {
	StoreID  *string `uri:"store_id"`
	DeviceID *string `uri:"device_id"`
}

type insertCoinToStoreCoinAcceptorRequest struct {
	Amount *int32 `json:"amount"`
}

func (s *Server) insertCoinsToStoreCoinAcceptor(c *gin.Context) {
	var reqUri insertCoinToStoreCoinAcceptorUri
	if err := c.ShouldBindUri(&reqUri); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqUri.StoreID == nil || *reqUri.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if reqUri.DeviceID == nil || *reqUri.DeviceID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "device_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*reqUri.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *reqUri.StoreID, *reqUri.DeviceID)))
		return
	}

	var reqJson insertCoinToStoreCoinAcceptorRequest
	if err := c.ShouldBindJSON(&reqJson); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqJson.Amount == nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "amount is null"))
		return
	}

	if *reqJson.Amount <= 0 {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "amount is smaller than or equal to 0"))
		return
	}

	// TODO: 檢查 amount 為某數倍數？

	arg1 := db.GetStoreDeviceParams{
		StoreID:  storeID,
		DeviceID: *reqUri.DeviceID,
	}

	if _, err := s.store.GetStoreDevice(c, arg1); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreDeviceNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *reqUri.StoreID, *reqUri.DeviceID)))
			return
		}
		logutil.GetLogger().Errorf("get store device error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)
	userID := authPayload.Subject
	var balanceEarmarkAmount, pointsEarmarkAmount int32

	{
		m := s.rs.NewMutex(distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
		if err := m.Lock(); err != nil {
			logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}

		unlock := func() {
			if ok, err := m.Unlock(); !ok || err != nil {
				logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
			}
		}

		arg2 := db.GetStoreUserParams{
			StoreID: storeID,
			UserID:  userID,
		}

		storeUser, err := s.store.GetStoreUser(c, arg2)
		if err != nil {
			logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg2)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			unlock()
			return
		}

		scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
		if !contains(scopes, roleutil.ScopeStoreDeviceInsertCoinsWithNegativeBalance) {
			if storeUser.Balance+storeUser.Points < *reqJson.Amount {
				c.JSON(http.StatusBadRequest, newErrorResponse(codeLowBalanceError, fmt.Sprintf("low balance, balance=%d, points=%d, amount=%d", storeUser.Balance, storeUser.Points, *reqJson.Amount)))
				unlock()
				return
			}
		}

		amount := *reqJson.Amount

		if storeUser.Points > amount {
			pointsEarmarkAmount = amount
			amount = 0
		} else {
			pointsEarmarkAmount = storeUser.Points
			amount -= storeUser.Points
		}

		balanceEarmarkAmount = amount

		arg3 := db.SetStoreUserBalanceWithLogParams{
			ChangedAt:        time.Now().UnixMilli(),
			ChangeType:       storeUserChangedTypeUpdateBalance,
			ChangedBy:        uuid.NullUUID{Valid: true, UUID: userID},
			ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
			ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
			StoreID:          storeID,
			UserID:           userID,
			Balance:          storeUser.Balance - balanceEarmarkAmount,
			Points:           storeUser.Points - pointsEarmarkAmount,
			BalanceEarmark:   storeUser.BalanceEarmark + balanceEarmarkAmount,
			PointsEarmark:    storeUser.PointsEarmark + pointsEarmarkAmount,
		}

		if err := s.store.SetStoreUserBalanceWithLog(c, arg3); err != nil {
			logutil.GetLogger().Errorf("set store user balance with log error, err=%s, arg=%#v", err, arg3)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			unlock()
			return
		}

		unlock()
	}

	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()

	if err := s.iot.AddPointsToCoinAcceptor(ctx, storeID, *reqUri.DeviceID, *reqJson.Amount); err != nil {
		{
			m := s.rs.NewMutex(distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
			if err := m.Lock(); err != nil {
				logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
				c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
				return
			}

			defer func() {
				if ok, err := m.Unlock(); !ok || err != nil {
					logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
				}
			}()

			arg2 := db.GetStoreUserParams{
				StoreID: storeID,
				UserID:  userID,
			}

			storeUser, err := s.store.GetStoreUser(c, arg2)
			if err != nil {
				logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg2)
				c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
				return
			}

			arg3 := db.SetStoreUserBalanceWithLogParams{
				ChangedAt:        time.Now().UnixMilli(),
				ChangeType:       storeUserChangedTypeUpdateBalance,
				ChangedBy:        uuid.NullUUID{Valid: true, UUID: userID},
				ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
				ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
				StoreID:          storeID,
				UserID:           userID,
				Balance:          storeUser.Balance + balanceEarmarkAmount,
				Points:           storeUser.Points + pointsEarmarkAmount,
				BalanceEarmark:   storeUser.BalanceEarmark - balanceEarmarkAmount,
				PointsEarmark:    storeUser.PointsEarmark - pointsEarmarkAmount,
			}

			if err := s.store.SetStoreUserBalanceWithLog(c, arg3); err != nil {
				logutil.GetLogger().Errorf("set store user balance with log error, err=%s, arg=%#v", err, arg3)
				c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
				return
			}
		}
		switch err.(type) {
		case *iotsdk.DeviceNotFoundError:
			c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreDeviceNotOnlineError, fmt.Sprintf("store device is not online, store_id=%s, device_id=%s", *reqUri.StoreID, *reqUri.DeviceID)))
		case *iotsdk.StoreNotFoundError:
			c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreNotOnlineError, fmt.Sprintf("store is not online, store_id=%s", *reqUri.StoreID)))
		default:
			logutil.GetLogger().Errorf("get coin acceptor status error, err=%s, store_id=%s, device_id=%s", err, storeID, *reqUri.DeviceID)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		}
		return
	}

	{
		m := s.rs.NewMutex(distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
		if err := m.Lock(); err != nil {
			logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}

		defer func() {
			if ok, err := m.Unlock(); !ok || err != nil {
				logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), userID.String()))
			}
		}()

		arg2 := db.GetStoreUserParams{
			StoreID: storeID,
			UserID:  userID,
		}

		storeUser, err := s.store.GetStoreUser(c, arg2)
		if err != nil {
			logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg2)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}

		arg3 := db.SetStoreUserBalanceWithLogParams{
			ChangedAt:        time.Now().UnixMilli(),
			ChangeType:       storeUserChangedTypeUpdateBalance,
			ChangedBy:        uuid.NullUUID{Valid: true, UUID: userID},
			ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
			ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
			StoreID:          storeID,
			UserID:           userID,
			Balance:          storeUser.Balance,
			Points:           storeUser.Points,
			BalanceEarmark:   storeUser.BalanceEarmark - balanceEarmarkAmount,
			PointsEarmark:    storeUser.PointsEarmark - pointsEarmarkAmount,
		}

		if err := s.store.SetStoreUserBalanceWithLog(c, arg3); err != nil {
			logutil.GetLogger().Errorf("set store user balance with log error, err=%s, arg=%#v", err, arg3)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	arg := db.CreateRecordParams{
		CreatedBy:        uuid.NullUUID{Valid: true, UUID: userID},
		CreatedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		CreatedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		Type:             db.RecordTypeCoinAcceptorRemoteInsertCoins,
		StoreID:          storeID,
		UserID:           uuid.NullUUID{Valid: true, UUID: userID},
		DeviceID:         sql.NullString{Valid: true, String: *reqUri.DeviceID},
		Amount:           balanceEarmarkAmount,
		PointAmount:      sql.NullInt32{Valid: true, Int32: pointsEarmarkAmount},
		Ts:               time.Now().UnixMilli(),
	}

	if _, err := s.store.CreateRecord(c, arg); err != nil {
		logutil.GetLogger().Errorf("create record error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type getStoreDeviceRecordsUri struct {
	StoreID  *string `uri:"store_id"`
	DeviceID *string `uri:"device_id"`
}

func (s *Server) getStoreDeviceRecords(c *gin.Context) {
	var req getStoreDeviceRecordsUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.DeviceID == nil || *req.DeviceID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "device_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
		return
	}

	arg1 := db.GetStoreDeviceParams{
		StoreID:  storeID,
		DeviceID: *req.DeviceID,
	}

	if _, err := s.store.GetStoreDevice(c, arg1); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreDeviceNotFoundError, fmt.Sprintf("store device not found, store_id=%s, device_id=%s", *req.StoreID, *req.DeviceID)))
			return
		}
		logutil.GetLogger().Errorf("get store device error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg2 := db.GetStoreDeviceRecordsParams{
		StoreID:  storeID,
		DeviceID: *req.DeviceID,
	}

	storeDeviceRecords, err := s.store.GetStoreDeviceRecords(c, arg2)
	if err != nil {
		logutil.GetLogger().Errorf("get store device records error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	records := make([]gin.H, 0, len(storeDeviceRecords))
	for _, record := range storeDeviceRecords {
		switch record.Type {
		case db.RecordTypeCoinAcceptorCoinInserted:
			records = append(records, gin.H{
				"type":   record.Type,
				"amount": record.Amount,
				"ts":     record.Ts,
			})
		case db.RecordTypeCoinAcceptorRemoteInsertCoins:
			records = append(records, gin.H{
				"type":         record.Type,
				"user_id":      record.UserID,
				"user_name":    record.UserName.String,
				"amount":       record.Amount,
				"point_amount": record.PointAmount.Int32,
				"ts":           record.Ts,
			})
		default:
			logutil.GetLogger().Warnf("unknown store device record type error, store_id=%s, device_id=%s, type=%s", storeID, *req.DeviceID, record.Type)
		}
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}
