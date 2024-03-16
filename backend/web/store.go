package web

import (
	db "backend/db/sqlc"
	"backend/token"
	distlockutil "backend/util/distlock"
	fsmutil "backend/util/fsm"
	logutil "backend/util/log"
	passwordutil "backend/util/password"
	randomutil "backend/util/random"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	storeChangedTypeCreate      string = "create"
	storeChangedTypeEnable      string = "enable"
	storeChangedTypeDeactive    string = "deactive"
	storeChangedTypeUpdateInfo  string = "update_info"
	storeChangedTypeGenPassword string = "gen_password"
)

type createStoreRequest struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

func (s *Server) createStore(c *gin.Context) {
	var req createStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.Name == nil || *req.Name == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is null or empty"))
		return
	}

	if regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`).MatchString(*req.Name) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is invalid"))
		return
	}

	if len(*req.Name) > int(s.config.MaxStoreNameLength) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError,
			fmt.Sprintf("name longer than %d characters", s.config.MaxStoreNameLength)))
		return
	}

	if req.Address == nil || *req.Address == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "address is null or empty"))
		return
	}

	if regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`).MatchString(*req.Address) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is invalid"))
		return
	}

	if len(*req.Address) > int(s.config.MaxStoreAddressLength) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError,
			fmt.Sprintf("address longer than %d characters", s.config.MaxStoreAddressLength)))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateStoreWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeChangedTypeCreate,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               uuid.New(),
		Name:             *req.Name,
		Address:          *req.Address,
		State:            fsmutil.InitStoreState,
	}

	store, err := s.store.CreateStoreWithLog(c, arg)
	if err != nil {
		logutil.GetLogger().Errorf("create store with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      store.ID.String(),
		"name":    store.Name,
		"address": store.Address,
		"state":   store.State,
	})
}

func (s *Server) getStores(c *gin.Context) {
	stores, err := s.store.GetStores(c)
	if err != nil {
		logutil.GetLogger().Errorf("get stores error, err=%s", err)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	res := make([]gin.H, 0, len(stores))
	for _, store := range stores {
		res = append(res, gin.H{
			"id":      store.ID.String(),
			"name":    store.Name,
			"address": store.Address,
			"state":   store.State,
		})
	}
	c.JSON(http.StatusOK, gin.H{"stores": res})
}

type getStoreUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) getStore(c *gin.Context) {
	var req getStoreUri
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
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *req.StoreID)))
		return
	}

	store, err := s.store.GetStore(c, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *req.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      store.ID.String(),
		"name":    store.Name,
		"address": store.Address,
		"state":   store.State,
	})
}

type enableStoreUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) enableStore(c *gin.Context) {
	var req enableStoreUri
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
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *req.StoreID)))
		return
	}

	m := s.rs.NewMutex(distlockutil.GetStoreIDMutexName(*req.StoreID))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreIDMutexName(*req.StoreID))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreIDMutexName(*req.StoreID))
		}
	}()

	store, err := s.store.GetStore(c, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *req.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	storeFSM := fsmutil.NewStoreFSM(store.State)
	if err := storeFSM.Event(c, fsmutil.StoreEventEnable); err != nil {
		switch store.State {
		case fsmutil.StoreStateActive:
			c.Status(http.StatusNoContent)
			return
		default:
			logutil.GetLogger().Errorf("store fsm error, err=%s, init_state=%s, event=%s", err, store.State, fsmutil.StoreEventEnable)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.SetStoreStateWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeChangedTypeEnable,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               storeID,
		State:            storeFSM.Current(),
	}
	if err := s.store.SetStoreStateWithLog(c, arg); err != nil {
		logutil.GetLogger().Errorf("set store state with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type deactiveStoreUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) deactiveStore(c *gin.Context) {
	var req deactiveStoreUri
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
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *req.StoreID)))
		return
	}

	m := s.rs.NewMutex(distlockutil.GetStoreIDMutexName(*req.StoreID))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreIDMutexName(*req.StoreID))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreIDMutexName(*req.StoreID))
		}
	}()

	store, err := s.store.GetStore(c, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *req.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	storeFSM := fsmutil.NewStoreFSM(store.State)
	if err := storeFSM.Event(c, fsmutil.StoreEventDeactive); err != nil {
		switch store.State {
		case fsmutil.StoreStateArchived:
			c.Status(http.StatusNoContent)
			return
		default:
			logutil.GetLogger().Errorf("store fsm error, err=%s, init_state=%s, event=%s", err, store.State, fsmutil.StoreEventDeactive)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.SetStoreStateWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeChangedTypeDeactive,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               storeID,
		State:            storeFSM.Current(),
	}
	if err := s.store.SetStoreStateWithLog(c, arg); err != nil {
		logutil.GetLogger().Errorf("set store state with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type updateStoreInfoUri struct {
	StoreID *string `uri:"store_id"`
}

type updateStoreInfoRequest struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

func (s *Server) updateStoreInfo(c *gin.Context) {
	var reqJson updateStoreInfoRequest
	if err := c.ShouldBindJSON(&reqJson); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqJson.Name == nil || *reqJson.Name == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is null or empty"))
		return
	}

	if regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`).MatchString(*reqJson.Name) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is invalid"))
		return
	}

	if len(*reqJson.Name) > int(s.config.MaxStoreNameLength) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError,
			fmt.Sprintf("name longer than %d characters", s.config.MaxStoreNameLength)))
		return
	}

	if reqJson.Address == nil || *reqJson.Address == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "address is null or empty"))
		return
	}

	if regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`).MatchString(*reqJson.Address) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is invalid"))
		return
	}

	if len(*reqJson.Address) > int(s.config.MaxStoreAddressLength) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError,
			fmt.Sprintf("address longer than %d characters", s.config.MaxStoreAddressLength)))
		return
	}

	var reqUri updateStoreInfoUri
	if err := c.ShouldBindUri(&reqUri); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqUri.StoreID == nil || *reqUri.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*reqUri.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *reqUri.StoreID)))
		return
	}

	store, err := s.store.GetStore(c, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *reqUri.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	if *reqJson.Name == store.Name && *reqJson.Address == store.Address {
		c.Status(http.StatusNoContent)
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.SetStoreNameAndAddressWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeChangedTypeUpdateInfo,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               storeID,
		Name:             *reqJson.Name,
		Address:          *reqJson.Address,
	}

	if err := s.store.SetStoreNameAndAddressWithLog(c, arg); err != nil {
		logutil.GetLogger().Errorf("set name and address with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type genStorePasswordUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) genStorePassword(c *gin.Context) {
	var reqUri genStorePasswordUri
	if err := c.ShouldBindUri(&reqUri); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqUri.StoreID == nil || *reqUri.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*reqUri.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *reqUri.StoreID)))
		return
	}

	password := randomutil.RandomAlphaNumString(int(s.config.StorePasswordLength))
	passwordHashed, err := passwordutil.HashPassword(password)
	if err != nil {
		logutil.GetLogger().Errorf("hash password error, err=%s", err)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.SetStorePasswordWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeChangedTypeGenPassword,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               storeID,
		Password:         sql.NullString{Valid: true, String: passwordHashed},
	}

	if err := s.store.SetStorePasswordWithLog(c, arg); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *reqUri.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("set store password with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}
