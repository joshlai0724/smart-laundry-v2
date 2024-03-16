package web

import (
	db "backend/db/sqlc"
	iotsdk "backend/iot-sdk"
	"backend/token"
	configutil "backend/util/config"
	logutil "backend/util/log"
	roleutil "backend/util/role"
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
)

type Server struct {
	config configutil.Config

	router *gin.Engine
	srv    *http.Server

	store      db.IStore
	rs         *redsync.Redsync
	tokenMaker token.Maker
	iot        iotsdk.IoT
}

func New(config configutil.Config, store db.IStore, rs *redsync.Redsync, tokenMaker token.Maker, iot iotsdk.IoT) (*Server, error) {
	server := &Server{
		config:     config,
		store:      store,
		rs:         rs,
		tokenMaker: tokenMaker,
		iot:        iot,
	}
	server.setupRouter()
	return server, nil
}

func (s *Server) setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST"},
		AllowHeaders:    []string{"authorization", "content-type"},
	}))

	// TODO: rate limit middleware
	v1Router := router.Group("/v1")

	v1Router.POST("/users/send-check-phone-number-owner-msg", s.sendCheckPhoneNumberOwnerMsg)
	v1Router.POST("/users/send-reset-password-msg", s.sendResetPasswordMsg)
	v1Router.GET("/users/check-phone-number-owner", s.checkPhoneNumberOwner)
	v1Router.POST("/users/.register", s.registerUser)
	v1Router.POST("/users/login", s.loginUser)
	v1Router.POST("/users/renew-access-token", s.renewAccessToken)
	v1Router.POST("/users/.reset-password", s.resetUserPassword)

	v1UserAuthRoutes := v1Router.Group("/").Use(
		authMiddleware(s.tokenMaker, s.checkToken),
		userScopesMiddleware(s.store),
	)

	v1UserAuthRoutes.GET("/users/info", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeUserDataRead}), s.getUserInfo)
	v1UserAuthRoutes.GET("/users/stores", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeUserDataRead}), s.getUserStores)
	v1UserAuthRoutes.GET("/users/scopes", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeUserDataRead}), s.getUserScopes)
	v1UserAuthRoutes.POST("/users/update-self-info", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeUserDataWrite}), s.updateUserSelfInfo)
	v1UserAuthRoutes.POST("/users/.change-password", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeUserDataWrite}), s.changeUserPassword)

	v1UserAuthRoutes.POST("/stores/.create", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreCreate}), s.createStore)
	v1UserAuthRoutes.GET("/stores", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreRead}), s.getStores)
	v1UserAuthRoutes.GET("/stores/:store_id", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreRead}), s.getStore)
	v1UserAuthRoutes.POST("/stores/:store_id/.enable", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreEnable}), s.enableStore)
	v1UserAuthRoutes.POST("/stores/:store_id/.deactive", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeactive}), s.deactiveStore)
	v1UserAuthRoutes.POST("/stores/:store_id/update-info", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreWrite}), s.updateStoreInfo)
	v1UserAuthRoutes.POST("/stores/:store_id/gen-password", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStorePasswordWrite}), s.genStorePassword)

	v1UserAuthRoutes.POST("/stores/:store_id/users/.register", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserAdminRegister},
		roleutil.Scopes{roleutil.ScopeStoreUserHqRegister},
		roleutil.Scopes{roleutil.ScopeStoreUserCustRegister},
	), s.registerStoreUser)

	v1StoreUserAuthRoutes := v1Router.Group("/").Use(
		authMiddleware(s.tokenMaker, s.checkToken),
		storeUserScopesMiddleware(s.store),
	)

	v1StoreUserAuthRoutes.GET("/stores/:store_id/users/scopes", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreUserDataRead}), s.getStoreUserScopes)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/users/:user_id/.enable", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserOwnerEnable},
		roleutil.Scopes{roleutil.ScopeStoreUserMgrEnable},
		roleutil.Scopes{roleutil.ScopeStoreUserCustEnable},
	), s.enableStoreUser)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/users/:user_id/.deactive", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserOwnerDeactive},
		roleutil.Scopes{roleutil.ScopeStoreUserMgrDeactive},
		roleutil.Scopes{roleutil.ScopeStoreUserCustDeactive},
	), s.deactiveStoreUser)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/users/:user_id/cust-cash-top-up", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreUserCustCashTopUp}), s.assistCustCashTopUp)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/users/:user_id/change-to-owner", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserOwnerEnable, roleutil.ScopeStoreUserMgrDeactive},
		roleutil.Scopes{roleutil.ScopeStoreUserOwnerEnable, roleutil.ScopeStoreUserCustDeactive},
	), s.changeStoreUserToOwner)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/users/:user_id/change-to-mgr", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserMgrEnable, roleutil.ScopeStoreUserOwnerDeactive},
		roleutil.Scopes{roleutil.ScopeStoreUserMgrEnable, roleutil.ScopeStoreUserCustDeactive},
	), s.changeStoreUserToMgr)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/users/:user_id/change-to-cust", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserCustEnable, roleutil.ScopeStoreUserOwnerDeactive},
		roleutil.Scopes{roleutil.ScopeStoreUserCustEnable, roleutil.ScopeStoreUserMgrDeactive},
	), s.changeStoreUserToCust)
	v1StoreUserAuthRoutes.GET("/stores/:store_id/users/:user_id/balance", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserRecordsReadSelf},
		roleutil.Scopes{roleutil.ScopeStoreUserRecordsReadOthers},
	), s.getStoreUserBalance)
	v1StoreUserAuthRoutes.GET("/stores/:store_id/users/:user_id/records", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreUserRecordsReadSelf},
		roleutil.Scopes{roleutil.ScopeStoreUserRecordsReadOthers},
	), s.getStoreUserRecords)
	v1StoreUserAuthRoutes.GET("/stores/:store_id/users", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreUserRead}), s.getStoreUsers)

	v1StoreUserAuthRoutes.GET("/stores/:store_id/devices", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeviceRead}), s.getStoreDevices)
	v1StoreUserAuthRoutes.GET("/stores/:store_id/devices/:device_id/records", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeviceRecordsRead}), s.getStoreDeviceRecords)
	v1StoreUserAuthRoutes.GET("/stores/:store_id/coin-acceptors/:device_id/info", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeviceRead}), s.getStoreCoinAcceptorInfo)
	v1StoreUserAuthRoutes.GET("/stores/:store_id/coin-acceptors/:device_id/status", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeviceRead}), s.getStoreCoinAcceptorStatus)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/coin-acceptors/:device_id/blink", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeviceBlink}), s.blinkStoreCoinAcceptor)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/coin-acceptors/:device_id/update-info", checkScopesMiddleware(roleutil.Scopes{roleutil.ScopeStoreDeviceWrite}), s.updateStoreCoinAcceptorInfo)
	v1StoreUserAuthRoutes.POST("/stores/:store_id/coin-acceptors/:device_id/insert-coins", checkScopesMiddleware(
		roleutil.Scopes{roleutil.ScopeStoreDeviceInsertCoins},
		roleutil.Scopes{roleutil.ScopeStoreDeviceInsertCoinsWithNegativeBalance},
	), s.insertCoinsToStoreCoinAcceptor)

	s.router = router
}

func (s *Server) Start(address string) error {
	s.srv = &http.Server{
		Addr:    address,
		Handler: s.router,
	}
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) checkToken(c context.Context, id uuid.UUID) bool {
	dbToken, err := s.store.GetToken(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		logutil.GetLogger().Errorf("get token error, err=%s, token_id=%s", err, id)
		return false
	}

	if dbToken.Type != token.TypeAccess {
		return false
	}

	if dbToken.IsBlocked {
		return false
	}

	if time.Now().After(time.UnixMilli(dbToken.ExpiredAt)) {
		return false
	}

	return true
}
