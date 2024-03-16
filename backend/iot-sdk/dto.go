package iotsdk

type messageType1[T any] struct {
	Type    string `json:"type"`
	CorrID  string `json:"corr_id"`
	Request T      `json:"request"`
	Ts1     int64  `json:"ts1"`
}

type messageType2[T any] struct {
	Type     string `json:"type"`
	CorrID   string `json:"corr_id"`
	Response T      `json:"response"`
	Error    *Error `json:"error"`
	Ts2      int64  `json:"ts2"`
	Ts3      int64  `json:"ts3"`
}

func (m *messageType2[T]) Err() error {
	if m.Error == nil {
		return nil
	}
	switch m.Error.Code {
	case "InvalidParameterError":
		return &InvalidParameterError{S: m.Error.Message}
	case "DeviceNotFoundError":
		return &DeviceNotFoundError{S: m.Error.Message}
	case "IllegalStateError":
		return &IllegalStateError{S: m.Error.Message}
	case "InternalError":
		return &InternalError{S: m.Error.Message}
	}
	return &InternalError{S: m.Error.Message}
}

type messageType3[T any] struct {
	Type  string `json:"type"`
	Event T      `json:"event"`
	Ts3   int64  `json:"ts3"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
