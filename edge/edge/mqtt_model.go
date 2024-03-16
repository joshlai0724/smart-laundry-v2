package edge

type MqttEvent struct {
	DeviceID string `json:"device_id"`
	Amount   int32  `json:"amount"`
	Points   int32  `json:"points"`
	State    string `json:"state"`
	Ts       int64  `json:"ts"`
}
