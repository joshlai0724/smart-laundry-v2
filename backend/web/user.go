package web

import (
	db "backend/db/sqlc"
	"backend/token"
	distlockutil "backend/util/distlock"
	fsmutil "backend/util/fsm"
	logutil "backend/util/log"
	passwordutil "backend/util/password"
	randomutil "backend/util/random"
	roleutil "backend/util/role"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/looplab/fsm"
)

const (
	userChangedTypeCreate                  string = "create"
	userChangedTypePasswordError           string = "password_error"
	userChangedTypeResetPasswordErrorCount string = "reset_password_error_count"
	userChangedTypeResetPassword           string = "reset_password"
	userChangedTypeChangePassword          string = "change_password"
	userChangedTypeUpdateInfo              string = "update_info"
)

type sendCheckPhoneNumberOwnerMsgRequest struct {
	PhoneNumber *string `json:"phone_number"`
}

func (s *Server) sendCheckPhoneNumberOwnerMsg(c *gin.Context) {
	// TODO: rate limit
	var req sendCheckPhoneNumberOwnerMsgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.PhoneNumber == nil || *req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is null or empty"))
		return
	}

	if len(*req.PhoneNumber) != 10 {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is not 10 characters"))
		return
	}

	if !strings.HasPrefix(*req.PhoneNumber, "09") {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number should start from 09"))
		return
	}

	_, err := s.store.GetUserByPhoneNumber(c, *req.PhoneNumber)
	if err != nil {
		if err != sql.ErrNoRows {
			logutil.GetLogger().Errorf("get user by phone number error, err=%s, phone_number=%s", err, *req.PhoneNumber)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, newErrorResponse(codePhoneNumberRegisteredError, "phone number exists"))
		return
	}

	m := s.rs.NewMutex(distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
		}
	}()

	arg1 := db.GetVerCodesByTypeAndPhoneNumberParams{
		Type:        verCodeTypeCheckPhoneNumberOwner,
		PhoneNumber: *req.PhoneNumber,
		FromTs:      time.Now().Add(-s.config.VerCode.CheckPhoneNumberOwner.TimePeriod).UnixMilli(),
	}
	codes, err := s.store.GetVerCodesByTypeAndPhoneNumber(c, arg1)
	if err != nil {
		logutil.GetLogger().Errorf("get ver code by phone number and type error, err=%s, arg=%#v", arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	if len(codes) >= s.config.VerCode.CheckPhoneNumberOwner.MaxMsgPerTimePeriod {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeSendCheckPhoneNumberOwnerMsgMeetLimitError,
			fmt.Sprintf("verification SMS for this phone number has reached the maximum limit, phone_number=%s", *req.PhoneNumber)))
		return
	}

	newCode := randomutil.RandomNumString(s.config.VerCode.CheckPhoneNumberOwner.Length)

	// TODO: 呼叫簡訊業者 API 寄出 ver code

	arg2 := db.CreateVerCodeParams{
		ID:          uuid.New(),
		PhoneNumber: *req.PhoneNumber,
		Code:        newCode,
		Type:        verCodeTypeCheckPhoneNumberOwner,
		RequestID:   "", // TODO: 簡訊業者 API 回傳的 RequestID
		ExpiredAt:   time.Now().Add(s.config.VerCode.CheckPhoneNumberOwner.LiveTime).UnixMilli(),
	}

	if _, err = s.store.CreateVerCode(c, arg2); err != nil {
		logutil.GetLogger().Errorf("create ver code error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type sendResetPasswordMsgRequest struct {
	PhoneNumber *string `json:"phone_number"`
}

func (s *Server) sendResetPasswordMsg(c *gin.Context) {
	// TODO: rate limit
	var req sendResetPasswordMsgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.PhoneNumber == nil || *req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is null or empty"))
		return
	}

	if len(*req.PhoneNumber) != 10 {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is not 10 characters"))
		return
	}

	if !strings.HasPrefix(*req.PhoneNumber, "09") {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number should start from 09"))
		return
	}

	m := s.rs.NewMutex(distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
		}
	}()

	_, err := s.store.GetUserByPhoneNumber(c, *req.PhoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, newErrorResponse(codePhoneNumberNotRegisterError,
				fmt.Sprintf("phone number is not registered as a member yet, phone_number=%s", *req.PhoneNumber)))
			return
		}
		logutil.GetLogger().Errorf("get user by phone number error, err=%s, phone_number=%s", err, *req.PhoneNumber)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg1 := db.GetVerCodesByTypeAndPhoneNumberParams{
		Type:        verCodeTypeResetPassword,
		PhoneNumber: *req.PhoneNumber,
		FromTs:      time.Now().Add(-s.config.VerCode.ResetPassword.TimePeriod).UnixMilli(),
	}
	codes, err := s.store.GetVerCodesByTypeAndPhoneNumber(c, arg1)
	if err != nil {
		logutil.GetLogger().Errorf("get ver code by phone number and type error, err=%s, arg=%#v", arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	if len(codes) >= s.config.VerCode.ResetPassword.MaxMsgPerTimePeriod {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeSendCheckPhoneNumberOwnerMsgMeetLimitError,
			fmt.Sprintf("password reset SMS limit has been reached, phone_number=%s", *req.PhoneNumber)))
		return
	}

	newCode := randomutil.RandomAlphaNumString(s.config.VerCode.ResetPassword.Length)

	// TODO: 呼叫簡訊業者 API 寄出 ver code

	arg2 := db.CreateVerCodeParams{
		ID:          uuid.New(),
		PhoneNumber: *req.PhoneNumber,
		Code:        newCode,
		Type:        verCodeTypeResetPassword,
		RequestID:   "", // TODO: 簡訊業者 API 回傳的 RequestID
		ExpiredAt:   time.Now().Add(s.config.VerCode.ResetPassword.LiveTime).UnixMilli(),
	}

	if _, err = s.store.CreateVerCode(c, arg2); err != nil {
		logutil.GetLogger().Errorf("create ver code error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type checkPhoneNumberOwnerParams struct {
	PhoneNumber *string `form:"phone_number"`
	VerCode     *string `form:"ver_code"`
}

func (s *Server) checkPhoneNumberOwner(c *gin.Context) {
	var req checkPhoneNumberOwnerParams
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.PhoneNumber == nil || *req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is null or empty"))
		return
	}

	if req.VerCode == nil || *req.VerCode == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "ver_code is null or empty"))
		return
	}

	if len(*req.VerCode) != s.config.VerCode.CheckPhoneNumberOwner.Length {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError,
			fmt.Sprintf("verification code should be %d digits", s.config.VerCode.CheckPhoneNumberOwner.Length)))
		return
	}

	arg := db.GetVerCodesByTypeAndPhoneNumberAndCodeParams{
		Type:        verCodeTypeCheckPhoneNumberOwner,
		PhoneNumber: *req.PhoneNumber,
		Code:        *req.VerCode,
	}

	codes, err := s.store.GetVerCodesByTypeAndPhoneNumberAndCode(c, arg)
	if err != nil {
		logutil.GetLogger().Errorf("get ver codes by type, phone number and code error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	var validCode *db.VerCode
	for _, code := range codes {
		if isVerCodeValid(code) {
			validCode = &code
			break
		}
	}

	if validCode == nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWrongVerCodeError, "verification code is incorrect"))
		return
	}
	c.Status(http.StatusNoContent)
}

type registerUserParams struct {
	PhoneNumber *string `json:"phone_number"`
	Name        *string `json:"name"`
	Password    *string `json:"password"`
	VerCode     *string `json:"ver_code"`
}

func (s *Server) registerUser(c *gin.Context) {
	var req registerUserParams
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.PhoneNumber == nil || *req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is null or empty"))
		return
	}

	if req.Name == nil || *req.Name == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "name is null or empty"))
		return
	}

	if req.Password == nil || *req.Password == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "password is null or empty"))
		return
	}

	if err := passwordutil.DoesPasswordMeetRule(*req.Password); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWeakPasswordError, "password is weak"))
		return
	}

	if req.VerCode == nil || *req.VerCode == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "ver_code is null or empty"))
		return
	}

	m := s.rs.NewMutex(distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(*req.PhoneNumber))
		}
	}()

	// 檢查 phone number 是否被申請過
	if _, err := s.store.GetUserByPhoneNumber(c, *req.PhoneNumber); err != nil {
		if err != sql.ErrNoRows {
			logutil.GetLogger().Errorf("get user by phone number error, err=%s, phone_number=%s", err, *req.PhoneNumber)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, newErrorResponse(codePhoneNumberRegisteredError, "phone number exists"))
		return
	}

	// 檢查 ver code 是否合法
	arg1 := db.GetVerCodesByTypeAndPhoneNumberAndCodeParams{
		Type:        verCodeTypeCheckPhoneNumberOwner,
		PhoneNumber: *req.PhoneNumber,
		Code:        *req.VerCode,
	}

	codes, err := s.store.GetVerCodesByTypeAndPhoneNumberAndCode(c, arg1)
	if err != nil {
		logutil.GetLogger().Errorf("get ver codes by type, phone number and code error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	var validCode *db.VerCode
	for _, code := range codes {
		if isVerCodeValid(code) {
			validCode = &code
			break
		}
	}

	if validCode == nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWrongVerCodeError, "verification code is incorrect"))
		return
	}

	// block 合法的 ver code
	if err := s.store.BlockVerCodes(c, validCode.ID); err != nil {
		logutil.GetLogger().Errorf("block ver code error, err=%s, id=%s", err, validCode.ID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	hashedPassword, err := passwordutil.HashPassword(*req.Password)
	if err != nil {
		logutil.GetLogger().Errorf("hash password error, err=%s", err)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	userID := uuid.New()

	// 將使用者加入 db
	arg2 := db.CreateUserWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       userChangedTypeCreate,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: userID},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               userID,
		PhoneNumber:      *req.PhoneNumber,
		Name:             *req.Name,
		Password:         hashedPassword,
		RoleID:           roleutil.GetRoleByName(roleutil.RoleMember).ID,
		State:            fsmutil.InitUserState,
	}
	if _, err := s.store.CreateUserWithLog(c, arg2); err != nil {
		logutil.GetLogger().Errorf("create user with log error, err=%s, arg=%#v", err, arg2)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type loginUserParams struct {
	PhoneNumber *string `json:"phone_number"`
	Password    *string `json:"password"`
}

func (s *Server) loginUser(c *gin.Context) {
	var req loginUserParams
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.PhoneNumber == nil || *req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "phone_number is null or empty"))
		return
	}

	if req.Password == nil || *req.Password == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "password is null or empty"))
		return
	}

	user, err := s.store.GetUserByPhoneNumber(c, *req.PhoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, newErrorResponse(codePhoneNumberOrPasswordError, "phone number or password incorrect"))
			return
		}
		logutil.GetLogger().Errorf("get user by phone number error, err=%s, phone_number=%s", err, *req.PhoneNumber)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	userFSM := fsmutil.NewUserFSM(user.State)
	if err := userFSM.Event(c, fsmutil.UserEventLogin); err != nil {
		if _, ok := err.(fsm.NoTransitionError); !ok {
			switch user.State {
			case fsmutil.UserStateLocked:
				c.JSON(http.StatusBadRequest, newErrorResponse(codeAccountLockedError,
					fmt.Sprintf("account has reached %d incorrect password attempts, please reset password", s.config.MaxPasswordAttempts)))
				return
			default:
				logutil.GetLogger().Errorf("user fsm error, err=%s, init_state=%s, event=%s", err, user.State, fsmutil.UserEventLogin)
				c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
				return
			}
		}
	}

	if err := passwordutil.CheckPassword(*req.Password, user.Password); err != nil {
		arg := db.SetUserPasswordAndStateWithLogParams{
			ChangedAt:          time.Now().UnixMilli(),
			ChangeType:         userChangedTypePasswordError,
			ChangedBy:          uuid.NullUUID{Valid: true, UUID: user.ID},
			ChangedUserAgent:   sql.NullString{Valid: true, String: c.Request.UserAgent()},
			ChangedClientIp:    sql.NullString{Valid: true, String: c.ClientIP()},
			ID:                 user.ID,
			Password:           user.Password,
			PasswordErrorCount: user.PasswordErrorCount + 1,
			PasswordChangedAt:  user.PasswordChangedAt,
			State:              user.State,
		}
		if arg.PasswordErrorCount >= s.config.MaxPasswordAttempts {
			arg.State = fsmutil.UserStateLocked
		}
		if err := s.store.SetUserPasswordAndStateWithLog(c, arg); err != nil {
			logutil.GetLogger().Errorf("set user password and state with log error, err=%s, arg=%#v", err, arg)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
		c.JSON(http.StatusBadRequest, newErrorResponse(codePhoneNumberOrPasswordError, "phone number or password incorrect"))
		return
	} else {
		if user.PasswordErrorCount != 0 {
			arg := db.SetUserPasswordAndStateWithLogParams{
				ChangedAt:          time.Now().UnixMilli(),
				ChangeType:         userChangedTypeResetPasswordErrorCount,
				ChangedBy:          uuid.NullUUID{Valid: true, UUID: user.ID},
				ChangedUserAgent:   sql.NullString{Valid: true, String: c.Request.UserAgent()},
				ChangedClientIp:    sql.NullString{Valid: true, String: c.ClientIP()},
				ID:                 user.ID,
				Password:           user.Password,
				PasswordErrorCount: 0,
				PasswordChangedAt:  user.PasswordChangedAt,
				State:              user.State,
			}
			if err := s.store.SetUserPasswordAndStateWithLog(c, arg); err != nil {
				logutil.GetLogger().Errorf("set user password and state with log error, err=%s, arg=%#v", err, arg)
				c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
				return
			}
		}
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, s.config.Token.AccessTokenDuration)
	if err != nil {
		logutil.GetLogger().Errorf("create access token error, err=%s, user_id=%s", err, user.ID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, s.config.Token.RefreshTokenDuration)
	if err != nil {
		logutil.GetLogger().Errorf("create refresh token error, err=%s, user_id=%s", err, user.ID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg := db.CreateTokenParams{
		ID:        accessPayload.ID,
		Type:      token.TypeAccess,
		UserAgent: c.Request.UserAgent(),
		ClientIp:  c.ClientIP(),
		UserID:    accessPayload.Subject,
		ExpiredAt: accessPayload.ExpiredAt,
		IssuedAt:  accessPayload.IssuedAt,
	}
	if _, err := s.store.CreateToken(c, arg); err != nil {
		logutil.GetLogger().Errorf("create access token error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg = db.CreateTokenParams{
		ID:        refreshPayload.ID,
		Type:      token.TypeRefresh,
		UserAgent: c.Request.UserAgent(),
		ClientIp:  c.ClientIP(),
		UserID:    refreshPayload.Subject,
		ExpiredAt: refreshPayload.ExpiredAt,
		IssuedAt:  refreshPayload.IssuedAt,
	}
	if _, err := s.store.CreateToken(c, arg); err != nil {
		logutil.GetLogger().Errorf("create refresh token error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    authorizationTypeBearer,
	})
}

type renewAccessTokenParams struct {
	RefreshToken *string `json:"refresh_token"`
}

func (s *Server) renewAccessToken(c *gin.Context) {
	var req renewAccessTokenParams
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.RefreshToken == nil || *req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "refresh_token is null or empty"))
		return
	}

	refreshPayload, err := s.tokenMaker.VerifyToken(*req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidTokenError, "invalid refresh token"))
		return
	}

	dbToken, err := s.store.GetToken(c, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidTokenError, "invalid refresh token"))
			return
		}
		logutil.GetLogger().Errorf("get token failed, err=%s, token_id=%s", err, dbToken.ID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	if dbToken.Type != token.TypeRefresh || dbToken.IsBlocked || time.Now().After(time.UnixMilli(dbToken.ExpiredAt)) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidTokenError, "invalid refresh token"))
		return
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Subject, s.config.Token.AccessTokenDuration)
	if err != nil {
		logutil.GetLogger().Errorf("create access token error, err=%s, user_id=%s", err, refreshPayload.Subject)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg := db.CreateTokenParams{
		ID:        accessPayload.ID,
		Type:      token.TypeAccess,
		UserAgent: c.Request.UserAgent(),
		ClientIp:  c.ClientIP(),
		UserID:    accessPayload.Subject,
		ExpiredAt: accessPayload.ExpiredAt,
		IssuedAt:  accessPayload.IssuedAt,
	}
	if _, err := s.store.CreateToken(c, arg); err != nil {
		logutil.GetLogger().Errorf("create access token error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

type resetUserPasswordParams struct {
	NewPassword *string `json:"new_password"`
	VerCode     *string `json:"ver_code"`
}

func (s *Server) resetUserPassword(c *gin.Context) {
	var req resetUserPasswordParams
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.NewPassword == nil || *req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "new_password is null or empty"))
		return
	}

	if err := passwordutil.DoesPasswordMeetRule(*req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWeakPasswordError, "new password is weak"))
		return
	}

	if req.VerCode == nil || *req.VerCode == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "ver_code is null or empty"))
		return
	}

	m1 := s.rs.NewMutex(distlockutil.GetVerCodeMutexName(*req.VerCode))
	if err := m1.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetVerCodeMutexName(*req.VerCode))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m1.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetVerCodeMutexName(*req.VerCode))
		}
	}()

	// 檢查 ver code 是否合法
	arg1 := db.GetVerCodesByTypeAndCodeParams{
		Type: verCodeTypeResetPassword,
		Code: *req.VerCode,
	}

	codes, err := s.store.GetVerCodesByTypeAndCode(c, arg1)
	if err != nil {
		logutil.GetLogger().Errorf("get ver codes by type and code error, err=%s, arg=%#v", err, arg1)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	var validCode *db.VerCode
	for _, code := range codes {
		if isVerCodeValid(code) {
			validCode = &code
			break
		}
	}

	if validCode == nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWrongVerCodeError, "verification code is incorrect"))
		return
	}

	m2 := s.rs.NewMutex(distlockutil.GetUserPhoneNumberMutexName(validCode.PhoneNumber))
	if err := m2.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(validCode.PhoneNumber))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m2.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(validCode.PhoneNumber))
		}
	}()

	user, err := s.store.GetUserByPhoneNumber(c, validCode.PhoneNumber)
	if err != nil {
		logutil.GetLogger().Errorf("get user by phone number error, err=%s, phone_number=%s", err, validCode.PhoneNumber)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	userFSM := fsmutil.NewUserFSM(user.State)
	if err := userFSM.Event(c, fsmutil.UserEventResetPassword); err != nil {
		if _, ok := err.(fsm.NoTransitionError); !ok {
			logutil.GetLogger().Errorf("user fsm error, err=%s, init_state=%s, event=%s", err, user.State, fsmutil.UserEventResetPassword)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	if err := passwordutil.CheckPassword(*req.NewPassword, user.Password); err == nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeNewPasswordIsOldPasswordError, "new password is the same as old password"))
		return
	}

	// block 合法的 ver code
	if err := s.store.BlockVerCodes(c, validCode.ID); err != nil {
		logutil.GetLogger().Errorf("block ver code error, err=%s, id=%s", err, validCode.ID)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	hashedPassword, err := passwordutil.HashPassword(*req.NewPassword)
	if err != nil {
		logutil.GetLogger().Errorf("hash password error, err=%s", err)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg := db.SetUserPasswordAndStateWithLogParams{
		ChangedAt:          time.Now().UnixMilli(),
		ChangeType:         userChangedTypeResetPassword,
		ChangedBy:          uuid.NullUUID{Valid: true, UUID: user.ID},
		ChangedUserAgent:   sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:    sql.NullString{Valid: true, String: c.ClientIP()},
		ID:                 user.ID,
		Password:           hashedPassword,
		PasswordErrorCount: 0,
		PasswordChangedAt:  sql.NullInt64{Valid: true, Int64: time.Now().UnixMilli()},
		State:              userFSM.Current(),
	}
	if err := s.store.SetUserPasswordAndStateWithLog(c, arg); err != nil {
		logutil.GetLogger().Errorf("set user password and state with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) getUserInfo(c *gin.Context) {
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	user, err := s.store.GetUser(c, authPayload.Subject)
	if err != nil {
		logutil.GetLogger().Errorf("get user error, err=%s, user_id=%s", err, authPayload.Subject)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": user.Name})
}

func (s *Server) getUserStores(c *gin.Context) {
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	stores, err := s.store.GetUserStores(c, authPayload.Subject)
	if err != nil {
		logutil.GetLogger().Errorf("get user stores error, err=%s, user_id=%s", err, authPayload.Subject)
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

func (s *Server) getUserScopes(c *gin.Context) {
	scopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)
	c.JSON(http.StatusOK, gin.H{"scopes": scopes})
}

type updateUserSelfInfoParams struct {
	Name *string `json:"name"`
}

func (s *Server) updateUserSelfInfo(c *gin.Context) {
	var req updateUserSelfInfoParams
	if err := c.ShouldBind(&req); err != nil {
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

	if len(*req.Name) > int(s.config.MaxUserNameLength) {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError,
			fmt.Sprintf("name longer than %d characters", s.config.MaxUserNameLength)))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.SetUserNameWithLogParams{
		ChangedAt:        time.Now().UnixMilli(),
		ChangeType:       userChangedTypeUpdateInfo,
		ChangedBy:        uuid.NullUUID{Valid: true, UUID: authPayload.Subject},
		ChangedUserAgent: sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:  sql.NullString{Valid: true, String: c.ClientIP()},
		ID:               authPayload.Subject,
		Name:             *req.Name,
	}

	if err := s.store.SetUserNameWithLog(c, arg); err != nil {
		logutil.GetLogger().Errorf("set user name with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}

type changeUserPasswordParams struct {
	OldPassword *string `json:"old_password"`
	NewPassword *string `json:"new_password"`
}

func (s *Server) changeUserPassword(c *gin.Context) {
	var req changeUserPasswordParams
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
		return
	}

	if req.OldPassword == nil || *req.OldPassword == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "old password is null or empty"))
		return
	}

	if req.NewPassword == nil || *req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "new password is null or empty"))
		return
	}

	if err := passwordutil.DoesPasswordMeetRule(*req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWeakPasswordError, "new password is weak"))
		return
	}

	if *req.NewPassword == *req.OldPassword {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeNewPasswordIsOldPasswordError, "new password is the same as old password"))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	user, err := s.store.GetUser(c, authPayload.Subject)
	if err != nil {
		logutil.GetLogger().Errorf("get user error, err=%s, user=%s", err, authPayload.Subject)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	phoneNumber := user.PhoneNumber

	m := s.rs.NewMutex(distlockutil.GetUserPhoneNumberMutexName(phoneNumber))
	if err := m.Lock(); err != nil {
		logutil.GetLogger().Errorf("lock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(phoneNumber))
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}
	defer func() {
		if ok, err := m.Unlock(); !ok || err != nil {
			logutil.GetLogger().Errorf("unlock error, err=%s, mutex_name=%s", err, distlockutil.GetUserPhoneNumberMutexName(phoneNumber))
		}
	}()

	user, err = s.store.GetUserByPhoneNumber(c, phoneNumber)
	if err != nil {
		logutil.GetLogger().Errorf("get user by phone number error, err=%s, phone_number=%s", err, phoneNumber)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	userFSM := fsmutil.NewUserFSM(user.State)
	if err := userFSM.Event(c, fsmutil.UserEventChangePassword); err != nil {
		if _, ok := err.(fsm.NoTransitionError); !ok {
			logutil.GetLogger().Errorf("user fsm error, err=%s, init_state=%s, event=%s", err, user.State, fsmutil.UserEventChangePassword)
			c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}
	}

	if err := passwordutil.CheckPassword(*req.OldPassword, user.Password); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(codeWrongOldPasswordError, "old password is incorrect"))
		return
	}

	hashedPassword, err := passwordutil.HashPassword(*req.NewPassword)
	if err != nil {
		logutil.GetLogger().Errorf("hash password error, err=%s", err)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	arg := db.SetUserPasswordAndStateWithLogParams{
		ChangedAt:          time.Now().UnixMilli(),
		ChangeType:         userChangedTypeChangePassword,
		ChangedBy:          uuid.NullUUID{Valid: true, UUID: user.ID},
		ChangedUserAgent:   sql.NullString{Valid: true, String: c.Request.UserAgent()},
		ChangedClientIp:    sql.NullString{Valid: true, String: c.ClientIP()},
		ID:                 user.ID,
		Password:           hashedPassword,
		PasswordErrorCount: 0,
		PasswordChangedAt:  sql.NullInt64{Valid: true, Int64: time.Now().UnixMilli()},
		State:              userFSM.Current(),
	}
	if err := s.store.SetUserPasswordAndStateWithLog(c, arg); err != nil {
		logutil.GetLogger().Errorf("set user password and state with log error, err=%s, arg=%#v", err, arg)
		c.JSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
		return
	}

	c.Status(http.StatusNoContent)
}
