package iot

type WsRequest struct {
	StoreID  *string `json:"store_id"`
	Password *string `json:"password"`
	DeviceID *string `json:"device_id"`
	RecordID *string `json:"record_id"`
	Amount   *int32  `json:"amount"`
	Ts       *int64  `json:"ts"`
}

type WsResponse struct {
	EdgeVersion     string `json:"edge_version"`
	FirmwareVersion string `json:"firmware_version"`
	Points          int32  `json:"points"`
	State           string `json:"state"`
	Ts              int64  `json:"ts"`
}

type WsEvent struct {
	DeviceID string `json:"device_id"`
	Points   int32  `json:"points"`
	State    string `json:"state"`
	Ts       int64  `json:"ts"`
}
