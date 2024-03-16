package edge

type Message[T any, U any] struct {
	Type     string `json:"type"`
	CorrID   string `json:"corr_id"`
	Request  *T     `json:"request"`
	Response *U     `json:"response,omitempty"`
	Error    *Error `json:"error,omitempty"`
	Ts1      int64  `json:"ts1"`
	Ts2      int64  `json:"ts2"`
	Ts3      int64  `json:"ts3"`
}

type MessageType1[T any] struct {
	Type    string `json:"type"`
	CorrID  string `json:"corr_id"`
	Request T      `json:"request"`
	Ts1     int64  `json:"ts1"`
}

type MessageType2[T any] struct {
	Type     string `json:"type"`
	CorrID   string `json:"corr_id"`
	Response T      `json:"response,omitempty"`
	Error    *Error `json:"error,omitempty"`
	Ts2      int64  `json:"ts2"`
	Ts3      int64  `json:"ts3"`
}

func (m *MessageType2[T]) Err() error {
	if m.Error == nil {
		return nil
	}
	switch m.Error.Code {
	case codeInvalidParameterError:
		return &InvalidParameterError{S: m.Error.Message}
	case codeDeviceNotFoundError:
		return &DeviceNotFoundError{S: m.Error.Message}
	case codeIllegalStateError:
		return &IllegalStateError{S: m.Error.Message}
	case codeInternalError:
		return &InternalError{S: m.Error.Message}
	}
	return &InternalError{S: m.Error.Message}
}

const (
	codeInvalidParameterError string = "InvalidParameterError"
	codeDeviceNotFoundError   string = "DeviceNotFoundError"
	codeIllegalStateError     string = "IllegalStateError"
	codeInternalError         string = "InternalError"
)

type InvalidParameterError struct {
	S string
}

func (e *InvalidParameterError) Error() string {
	return e.S
}

type NotFoundError struct {
	S string
}

func (e *NotFoundError) Error() string {
	return e.S
}

type DeviceNotFoundError struct {
	S string
}

func (e *DeviceNotFoundError) Error() string {
	return e.S
}

type InternalError struct {
	S string
}

func (e *InternalError) Error() string {
	return e.S
}

type IllegalStateError struct {
	S string
}

func (e *IllegalStateError) Error() string {
	return e.S
}

type MessageType3[T any] struct {
	Type  string `json:"type"`
	Event T      `json:"event,omitempty"`
	Ts3   int64  `json:"ts3"`
}

type CoinAcceptorInfo struct {
	FirmwareVersion string `json:"firmware_version"`
}

type CoinAcceptorStatus struct {
	Points int32  `json:"points"`
	State  string `json:"state"`
	Ts     int64  `json:"ts"`
}
