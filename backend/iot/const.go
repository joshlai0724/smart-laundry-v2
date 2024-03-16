package iot

const (
	// rbmq
	Exchange         string = "iot-backend"
	RequestKey       string = "api.v1.request"
	ResponseKeyFmt   string = "api.v1.response.%s"
	StoreEventKeyFmt string = "api.v1.event.store.%s"
)
