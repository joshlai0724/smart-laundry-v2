package edge

type WsRequest struct {
	DeviceID *string `json:"device_id"`
	Amount   *int32  `json:"amount"`
}

type WsResponse struct {
}
