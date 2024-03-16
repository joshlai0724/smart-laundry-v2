package web

func newErrorResponse(code, message string) map[string]string {
	return map[string]string{
		"code":    code,
		"message": message,
	}
}

const (
	messageWrongRequestPayload string = "request payload cannot be parsed"
	messageServerInternalError string = "internal error"
	messageForbiddenError      string = "no permission"

	codeInvalidParameterError                      string = "InvalidParameterError"
	codeInternalError                              string = "InternalError"
	codeForbiddenError                             string = "ForbiddenError"
	codeSendCheckPhoneNumberOwnerMsgMeetLimitError string = "SendCheckPhoneNumberOwnerMsgMeetLimitError"
	codeSendResetPasswordMsgMeetLimitError         string = "SendResetPasswordMsgMeetLimitError"
	codePhoneNumberRegisteredError                 string = "PhoneNumberRegisteredError"
	codePhoneNumberNotRegisterError                string = "PhoneNumberNotRegisterError"
	codeWrongVerCodeError                          string = "WrongVerCodeError"
	codeWeakPasswordError                          string = "WeakPasswordError"
	codeNewPasswordIsOldPasswordError              string = "NewPasswordIsOldPasswordError"
	codeWrongOldPasswordError                      string = "WrongOldPasswordError"
	codePhoneNumberOrPasswordError                 string = "PhoneNumberOrPasswordError"
	codeAccountLockedError                         string = "AccountLockedError"
	codeInvalidTokenError                          string = "InvalidTokenError"
	codeStoreUserRegisteredError                   string = "StoreUserRegisteredError"
	codeStoreUserNotRegisterError                  string = "StoreUserNotRegisterError"
	codeLowBalanceError                            string = "LowBalanceError"

	codeStoreNotFoundError       string = "StoreNotFoundError"
	codeStoreUserNotFoundError   string = "StoreUserNotFoundError"
	codeStoreDeviceNotFoundError string = "StoreDeviceNotFoundError"

	codeStoreDeviceNotOnlineError string = "StoreDeviceNotOnlineError"
	codeStoreNotOnlineError       string = "StoreNotOnlineError"
)
