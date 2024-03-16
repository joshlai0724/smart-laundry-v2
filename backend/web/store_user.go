package web

import (
	db "backend/db/sqlc"
	"backend/token"
	distlockutil "backend/util/distlock"
	fsmutil "backend/util/fsm"
	logutil "backend/util/log"
	roleutil "backend/util/role"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/looplab/fsm"
)

const (
	storeUserChangedTypeCreate        string = "create"
	storeUserChangedTypeEnable        string = "enable"
	storeUserChangedTypeDeactive      string = "deactive"
	storeUserChangedTypeChangeToOwner string = "change_to_owner"
	storeUserChangedTypeChangeToMgr   string = "change_to_mgr"
	storeUserChangedTypeChangeToCust  string = "change_to_cust"
	storeUserChangedTypeUpdateBalance string = "update_balance"
	storeUserChangedTypeCashTopUp     string = "cash_top_up"
)

type registerStoreUserUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) registerStoreUser(c *gin.Context) {
	var reqUri registerStoreUserUri
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

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	m := s.rs.NewMutex(distlockutil.GetStoreUserIDMutexName(storeID.String(), authPayload.Subject.String()))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), authPayload.Subject.String()))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetStoreUserIDMutexName(storeID.String(), authPayload.Subject.String()))
		}
	}()

	if _, err := s.store.GetStore(c, storeID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store is not, store_id=%s", *reqUri.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID.String())
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  authPayload.Subject,
	}

	if _, err := s.store.GetStoreUser(c, arg1); err != nil {
		if err != sql.ErrNoRows {
			logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeStoreUserRegisteredError, "store user exists"))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	roleName := roleutil.RoleCust
	if contains(scopes, roleutil.ScopeStoreUserAdminRegister) {
		roleName = roleutil.RoleAdmin
	} else if contains(scopes, roleutil.ScopeStoreUserHqRegister) {
		roleName = roleutil.RoleHq
	} else if contains(scopes, roleutil.ScopeStoreUserCustRegister) {
		roleName = roleutil.RoleCust
	}

	arg2 := db.CreateStoreUserWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeCreate,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           authPayload.Subject,
		RoleID:           roleutil.GetRoleByName(roleName).ID,
		State:            fsmutil.InitStoreUserState,
	}

	if _, err := s.store.CreateStoreUserWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("create store user with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) getStoreUserScopes(c *gin.Context) {
	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	c.JSON(http.StatusOK, gin.H{"scopes": scopes})
}

type enableStoreUserUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

func (s *Server) enableStoreUser(c *gin.Context) {
	var req enableStoreUserUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	userID, err := uuid.Parse(*req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

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

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !((roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleOwner && contains(scopes, roleutil.ScopeStoreUserOwnerEnable)) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleMgr && contains(scopes, roleutil.ScopeStoreUserMgrEnable) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleCust && contains(scopes, roleutil.ScopeStoreUserCustEnable)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	storeUserFSM := fsmutil.NewStoreUserFSM(storeUser.State)
	if err := storeUserFSM.Event(c, fsmutil.StoreUserEventEnable); err != nil {
		switch storeUser.State {
		case fsmutil.StoreUserStateActive:
			c.Status(http.StatusNoContent)
			return
		default:
			logutil.GetLogger().Errorf("store user fsm error, err=%s, init_state=%s, event=%s", err, storeUser.State, fsmutil.StoreUserEventEnable)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg2 := db.SetStoreUserStateWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeEnable,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           userID,
		State:            storeUserFSM.Current(),
	}
	if err := s.store.SetStoreUserStateWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store user state with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type deactiveStoreUserUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

func (s *Server) deactiveStoreUser(c *gin.Context) {
	var req deactiveStoreUserUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	userID, err := uuid.Parse(*req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

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

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !(roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleOwner && contains(scopes, roleutil.ScopeStoreUserOwnerDeactive) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleMgr && contains(scopes, roleutil.ScopeStoreUserMgrDeactive) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleCust && contains(scopes, roleutil.ScopeStoreUserCustDeactive)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	storeUserFSM := fsmutil.NewStoreUserFSM(storeUser.State)
	if err := storeUserFSM.Event(c, fsmutil.StoreUserEventDeactive); err != nil {
		switch storeUser.State {
		case fsmutil.StoreStateArchived:
			c.Status(http.StatusNoContent)
			return
		default:
			logutil.GetLogger().Errorf("store user fsm error, err=%s, init_state=%s, event=%s", err, storeUser.State, fsmutil.StoreUserEventDeactive)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg2 := db.SetStoreUserStateWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeDeactive,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           userID,
		State:            storeUserFSM.Current(),
	}
	if err := s.store.SetStoreUserStateWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store user state with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type changeStoreUserToOwnerUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

func (s *Server) changeStoreUserToOwner(c *gin.Context) {
	var req changeStoreUserToOwnerUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	userID, err := uuid.Parse(*req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

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

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !(roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleMgr && contains(scopes, roleutil.ScopeStoreUserMgrDeactive) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleCust && contains(scopes, roleutil.ScopeStoreUserCustDeactive)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg2 := db.SetStoreUserRoleIDWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeChangeToOwner,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           userID,
		RoleID:           roleutil.GetRoleByName(roleutil.RoleOwner).ID,
	}
	if err := s.store.SetStoreUserRoleIDWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store user role id with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type changeStoreUserToMgrUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

func (s *Server) changeStoreUserToMgr(c *gin.Context) {
	var req changeStoreUserToMgrUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	userID, err := uuid.Parse(*req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

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

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !(roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleOwner && contains(scopes, roleutil.ScopeStoreUserOwnerDeactive) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleCust && contains(scopes, roleutil.ScopeStoreUserCustDeactive)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg2 := db.SetStoreUserRoleIDWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeChangeToMgr,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           userID,
		RoleID:           roleutil.GetRoleByName(roleutil.RoleMgr).ID,
	}
	if err := s.store.SetStoreUserRoleIDWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store user role id with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type changeStoreUserToCustUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

func (s *Server) changeStoreUserToCust(c *gin.Context) {
	var req changeStoreUserToCustUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	userID, err := uuid.Parse(*req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

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

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !(roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleOwner && contains(scopes, roleutil.ScopeStoreUserOwnerDeactive) ||
		roleutil.GetRoleByID(storeUser.RoleID).Name == roleutil.RoleMgr && contains(scopes, roleutil.ScopeStoreUserMgrDeactive)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg2 := db.SetStoreUserRoleIDWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeChangeToCust,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           userID,
		RoleID:           roleutil.GetRoleByName(roleutil.RoleCust).ID,
	}
	if err := s.store.SetStoreUserRoleIDWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store user role id with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type getStoreUserBalanceUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

func (s *Server) getStoreUserBalance(c *gin.Context) {
	var req getStoreUserBalanceUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.StoreID == nil || *req.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*req.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	userID, err := uuid.Parse(*req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
		return
	}

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *req.StoreID, *req.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !(authPayload.Subject == arg1.UserID && contains(scopes, roleutil.ScopeStoreUserRecordsReadSelf) ||
		contains(scopes, roleutil.ScopeStoreUserRecordsReadOthers)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":         storeUser.Balance,
		"points":          storeUser.Points,
		"balance_earmark": storeUser.BalanceEarmark,
		"points_earmark":  storeUser.PointsEarmark,
	})
}

type getStoreUserRecordsUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

type getStoreUserRecordsQuery struct {
	Type *string `form:"type"`
}

func (s *Server) getStoreUserRecords(c *gin.Context) {
	var reqUri getStoreUserRecordsUri
	if err := c.ShouldBindUri(&reqUri); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqUri.StoreID == nil || *reqUri.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if reqUri.UserID == nil || *reqUri.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*reqUri.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
		return
	}

	userID, err := uuid.Parse(*reqUri.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
		return
	}

	var reqQuery getStoreUserRecordsQuery
	if err := c.ShouldBindQuery(&reqQuery); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqQuery.Type == nil || *reqQuery.Type == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "type is null or empty"))
		return
	}

	types := dtoRecordType2DbRecordType(*reqQuery.Type)
	if len(types) == 0 {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, fmt.Sprintf("type invalid, type=%s", *reqQuery.Type)))
		return
	}

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	if _, err := s.store.GetStoreUser(c, arg1); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	if !(authPayload.Subject == arg1.UserID && contains(scopes, roleutil.ScopeStoreUserRecordsReadSelf) ||
		contains(scopes, roleutil.ScopeStoreUserRecordsReadOthers)) {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
		return
	}

	arg2 := db.GetStoreUserRecordsParams{
		StoreID: storeID,
		UserID:  userID,
		Types:   types,
	}

	storeUserRecords, err := s.store.GetStoreUserRecords(c, arg2)
	if err != nil {
		logutil.GetLogger().Errorf("get store user records error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	records := make([]gin.H, 0, len(storeUserRecords))
	for _, record := range storeUserRecords {
		switch record.Type {
		case db.RecordTypeCashTopUp:
			records = append(records, gin.H{
				"type":                 record.Type,
				"created_by_user_id":   record.CreatedByUserID.UUID,
				"created_by_user_name": record.CreatedByUserName.String,
				"user_id":              record.UserID.UUID,
				"user_name":            record.UserName.String,
				"amount":               record.Amount,
				"points_amount":        record.PointAmount.Int32,
				"ts":                   record.Ts,
			})
		case db.RecordTypeCoinAcceptorRemoteInsertCoins:
			records = append(records, gin.H{
				"type":                record.Type,
				"device_id":           record.DeviceID.String,
				"device_name":         record.DeviceName.String,
				"device_real_type":    record.DeviceRealType.String,
				"device_display_type": record.DeviceDisplayType.String,
				"amount":              record.Amount,
				"point_amount":        record.PointAmount.Int32,
				"ts":                  record.Ts,
			})
		default:
			logutil.GetLogger().Warnf("unknown store user record type error, store_id=%s, user_id=%s, type=%s", storeID, userID, record.Type)
		}
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

type getStoreUsersUri struct {
	StoreID *string `uri:"store_id"`
}

func (s *Server) getStoreUsers(c *gin.Context) {
	var req getStoreUsersUri
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

	if _, err := s.store.GetStore(c, storeID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreNotFoundError, fmt.Sprintf("store not found, store_id=%s", *req.StoreID)))
			return
		}
		logutil.GetLogger().Errorf("get store error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	storeUsers, err := s.store.GetStoreUsersByStoreID(c, storeID)
	if err != nil {
		logutil.GetLogger().Errorf("get store users by store id error, err=%s, store_id=%s", err, storeID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	res := make([]gin.H, 0, len(storeUsers))
	for _, storeUser := range storeUsers {
		res = append(res, gin.H{
			"id":           storeUser.ID.String(),
			"phone_number": storeUser.PhoneNumber,
			"name":         storeUser.Name,
			"state":        storeUser.State,
			"role":         roleutil.GetRoleByID(storeUser.RoleID).Name,
		})
	}
	c.JSON(http.StatusOK, gin.H{"users": res})
}

type assistCustCashTopUpUri struct {
	StoreID *string `uri:"store_id"`
	UserID  *string `uri:"user_id"`
}

type assistCustCashTopUpRequest struct {
	Amount *int32 `uri:"amount"`
}

func (s *Server) assistCustCashTopUp(c *gin.Context) {
	var reqUri assistCustCashTopUpUri
	if err := c.ShouldBindUri(&reqUri); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if reqUri.StoreID == nil || *reqUri.StoreID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
		return
	}

	if reqUri.UserID == nil || *reqUri.UserID == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "user_id is null or empty"))
		return
	}

	storeID, err := uuid.Parse(*reqUri.StoreID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
		return
	}

	userID, err := uuid.Parse(*reqUri.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
		return
	}

	var reqJson assistCustCashTopUpRequest
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

	arg1 := db.GetStoreUserParams{
		StoreID: storeID,
		UserID:  userID,
	}

	storeUser, err := s.store.GetStoreUser(c, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, newErrorResponse(codeStoreUserNotFoundError, fmt.Sprintf("store user not found, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
			return
		}
		logutil.GetLogger().Errorf("get store user error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	if roleutil.GetRoleByID(storeUser.RoleID).Name != roleutil.RoleCust {
		c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, fmt.Sprintf("store user is not a cust, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
		return
	}

	storeUserFSM := fsmutil.NewStoreUserFSM(storeUser.State)
	if err := storeUserFSM.Event(c, fsmutil.StoreUserEventCashTopUp); err != nil {
		switch err.(type) {
		case fsm.InvalidEventError:
			c.JSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, fmt.Sprintf("store user state is not active, store_id=%s, user_id=%s", *reqUri.StoreID, *reqUri.UserID)))
			return
		case fsm.NoTransitionError:
		default:
			logutil.GetLogger().Errorf("store user fsm error, err=%s, init_state=%s, event=%s", err, storeUser.State, fsmutil.StoreUserEventCashTopUp)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	// TODO: 是否要考慮點數回饋？
	arg2 := db.SetStoreUserBalanceWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       storeUserChangedTypeCashTopUp,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		StoreID:          storeID,
		UserID:           userID,
		Balance:          storeUser.Balance + *reqJson.Amount,
		Points:           storeUser.Points,
		BalanceEarmark:   storeUser.BalanceEarmark,
		PointsEarmark:    storeUser.PointsEarmark,
	}

	if err := s.store.SetStoreUserBalanceWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("set store user balance with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg3 := db.CreateRecordParams{
		CreatedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		CreatedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		CreatedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		Type:             db.RecordTypeCashTopUp,
		StoreID:          storeID,
		UserID:           uuid.NullUUID{Valid: true, UUID: userID},
		Amount:           *reqJson.Amount,
		PointAmount:      sql.NullInt32{Valid: true, Int32: 0},
		Ts:               time.Now().UnixMilli(),
	}

	if _, err := s.store.CreateRecord(c, arg3); err != nil {
		logutil.GetLogger().Errorf("create record error, err=%s, arg=%#v", err, arg3)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}
