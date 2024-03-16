package distlockutil

var prefix = "mutex:"

func GetEdgeStoreIDMutexName(storeID string) string {
	return prefix + "edge-store-id:" + storeID
}

func GetUserPhoneNumberMutexName(phoneNumber string) string {
	return prefix + "user-phone-number:" + phoneNumber
}

func GetVerCodeMutexName(code string) string {
	return prefix + "ver-code:" + code
}

func GetStoreIDMutexName(storeID string) string {
	return prefix + "store-id:" + storeID
}

func GetStoreUserIDMutexName(storeID string, userID string) string {
	return prefix + "store-user-id:" + storeID + "+" + userID
}
