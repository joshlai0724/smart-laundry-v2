package roleutil

const (
	// user scope
	ScopeUserDataWrite          = "user:data:write"
	ScopeUserDataRead           = "user:data:read"
	ScopeStoreRead              = "store:read"
	ScopeStoreCreate            = "store:create"
	ScopeStoreEnable            = "store:enable"
	ScopeStoreDeactive          = "store:deactive"
	ScopeStoreWrite             = "store:write"
	ScopeStorePasswordWrite     = "store:password:write"
	ScopeStoreReportRead        = "store:report:read"
	ScopeStoreUserAdminRegister = "store:user:admin:register"
	ScopeStoreUserHqRegister    = "store:user:hq:register"
	ScopeStoreUserCustRegister  = "store:user:cust:register"

	// store user scope
	ScopeStoreDevice_RecordsRead                   = "store:device-records:read"
	ScopeStoreUser_RecordsRead                     = "store:user-records:read"
	ScopeStoreUserDataRead                         = "store:user:data:read"
	ScopeStoreUserOwnerEnable                      = "store:user:owner:enable"
	ScopeStoreUserMgrEnable                        = "store:user:mgr:enable"
	ScopeStoreUserCustEnable                       = "store:user:cust:enable"
	ScopeStoreUserOwnerDeactive                    = "store:user:owner:deactive"
	ScopeStoreUserMgrDeactive                      = "store:user:mgr:deactive"
	ScopeStoreUserCustDeactive                     = "store:user:cust:deactive"
	ScopeStoreUserOnlineTopUp                      = "store:user:online-top-up"
	ScopeStoreUserCustCashTopUp                    = "store:user:cust-cash-top-up"
	ScopeStoreUserRecordsReadSelf                  = "store:user:records:read-self"
	ScopeStoreUserRecordsReadOthers                = "store:user:records:read-others"
	ScopeStoreUserRead                             = "store:user:read"
	ScopeStoreDeviceRead                           = "store:device:read"
	ScopeStoreDeviceWrite                          = "store:device:write"
	ScopeStoreDeviceBlink                          = "store:device:blink"
	ScopeStoreDeviceInsertCoins                    = "store:device:insert-coins"
	ScopeStoreDeviceInsertCoinsWithNegativeBalance = "store:device:insert-coins-with-negative-balance"
	ScopeStoreDeviceRecordsRead                    = "store:device:records:read"
)
