package roleutil

const (
	RoleAdmin  = "admin"
	RoleHq     = "hq"
	RoleMember = "member"
	RoleOwner  = "owner"
	RoleMgr    = "mgr"
	RoleCust   = "cust"
)

type Scopes = []string

type Role struct {
	ID              int16
	Name            string
	UserScopes      []string
	StoreUserScopes []string
}

var adminRole = Role{
	ID:   1,
	Name: RoleAdmin,
	UserScopes: []string{
		ScopeUserDataWrite,
		ScopeUserDataRead,
		ScopeStoreRead,
		ScopeStoreCreate,
		ScopeStoreEnable,
		ScopeStoreDeactive,
		ScopeStoreWrite,
		ScopeStorePasswordWrite,
		ScopeStoreReportRead,
		ScopeStoreUserAdminRegister,
	},
	StoreUserScopes: []string{
		ScopeStoreDevice_RecordsRead,
		ScopeStoreUser_RecordsRead,
		ScopeStoreUserDataRead,
		ScopeStoreUserOwnerEnable,
		ScopeStoreUserMgrEnable,
		ScopeStoreUserCustEnable,
		ScopeStoreUserOwnerDeactive,
		ScopeStoreUserMgrDeactive,
		ScopeStoreUserCustDeactive,
		ScopeStoreUserCustCashTopUp,
		ScopeStoreUserRecordsReadSelf,
		ScopeStoreUserRecordsReadOthers,
		ScopeStoreUserRead,
		ScopeStoreDeviceRead,
		ScopeStoreDeviceWrite,
		ScopeStoreDeviceBlink,
		ScopeStoreDeviceInsertCoinsWithNegativeBalance,
		ScopeStoreDeviceRecordsRead,
	},
}

var hqRole = Role{
	ID:   2,
	Name: RoleHq,
	UserScopes: []string{
		ScopeUserDataWrite,
		ScopeUserDataRead,
		ScopeStoreRead,
		ScopeStoreCreate,
		ScopeStoreEnable,
		ScopeStoreDeactive,
		ScopeStoreWrite,
		ScopeStoreReportRead,
		ScopeStoreUserHqRegister,
	},
	StoreUserScopes: []string{
		ScopeStoreDevice_RecordsRead,
		ScopeStoreUser_RecordsRead,
		ScopeStoreUserDataRead,
		ScopeStoreUserOwnerEnable,
		ScopeStoreUserMgrEnable,
		ScopeStoreUserCustEnable,
		ScopeStoreUserOwnerDeactive,
		ScopeStoreUserMgrDeactive,
		ScopeStoreUserCustDeactive,
		ScopeStoreUserCustCashTopUp,
		ScopeStoreUserRecordsReadSelf,
		ScopeStoreUserRecordsReadOthers,
		ScopeStoreUserRead,
		ScopeStoreDeviceRead,
		ScopeStoreDeviceWrite,
		ScopeStoreDeviceBlink,
		ScopeStoreDeviceInsertCoinsWithNegativeBalance,
		ScopeStoreDeviceRecordsRead,
	},
}

var memberRole = Role{
	ID:   3,
	Name: RoleMember,
	UserScopes: []string{
		ScopeUserDataWrite,
		ScopeUserDataRead,
		ScopeStoreRead,
		ScopeStoreUserCustRegister,
	},
	StoreUserScopes: []string{},
}

var ownerRole = Role{
	ID:         4,
	Name:       RoleOwner,
	UserScopes: []string{},
	StoreUserScopes: []string{
		ScopeStoreDevice_RecordsRead,
		ScopeStoreUser_RecordsRead,
		ScopeStoreUserDataRead,
		ScopeStoreUserMgrEnable,
		ScopeStoreUserCustEnable,
		ScopeStoreUserMgrDeactive,
		ScopeStoreUserCustDeactive,
		ScopeStoreUserCustCashTopUp,
		ScopeStoreUserRecordsReadSelf,
		ScopeStoreUserRecordsReadOthers,
		ScopeStoreUserRead,
		ScopeStoreDeviceRead,
		ScopeStoreDeviceWrite,
		ScopeStoreDeviceBlink,
		ScopeStoreDeviceInsertCoinsWithNegativeBalance,
		ScopeStoreDeviceRecordsRead,
	},
}

var mgrRole = Role{
	ID:         5,
	Name:       RoleMgr,
	UserScopes: []string{},
	StoreUserScopes: []string{
		ScopeStoreDevice_RecordsRead,
		ScopeStoreUser_RecordsRead,
		ScopeStoreUserDataRead,
		ScopeStoreUserCustEnable,
		ScopeStoreUserCustDeactive,
		ScopeStoreUserCustCashTopUp,
		ScopeStoreUserRecordsReadSelf,
		ScopeStoreUserRecordsReadOthers,
		ScopeStoreUserRead,
		ScopeStoreDeviceRead,
		ScopeStoreDeviceWrite,
		ScopeStoreDeviceBlink,
		ScopeStoreDeviceInsertCoinsWithNegativeBalance,
		ScopeStoreDeviceRecordsRead,
	},
}

var custRole = Role{
	ID:         6,
	Name:       RoleCust,
	UserScopes: []string{},
	StoreUserScopes: []string{
		ScopeStoreUserDataRead,
		ScopeStoreUserOnlineTopUp,
		ScopeStoreUserRecordsReadSelf,
		ScopeStoreDeviceRead,
		ScopeStoreDeviceInsertCoins,
	},
}

var roles = []Role{adminRole, hqRole, memberRole, ownerRole, mgrRole, custRole}

func GetRoleByName(name string) Role {
	for _, role := range roles {
		if role.Name == name {
			return role
		}
	}
	return Role{}
}

func GetRoleByID(id int16) Role {
	for _, role := range roles {
		if role.ID == id {
			return role
		}
	}
	return Role{}
}
