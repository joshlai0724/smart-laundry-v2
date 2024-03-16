package iot

type RbmqRequest struct {
	StoreID  *string `json:"store_id"`
	DeviceID *string `json:"device_id"`
	Amount   *int32  `json:"amount"`
}
