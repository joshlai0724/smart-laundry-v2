package simulator

type MessageType1 struct {
	Type    string  `json:"type"`
	CorrID  string  `json:"corr_id"`
	Request Request `json:"request"`
	Ts1     int64   `json:"ts1"`
}

type MessageType2 struct {
	Type     string `json:"type"`
	CorrID   string `json:"corr_id"`
	Response any    `json:"response,omitempty"`
	Error    *Error `json:"error,omitempty"`
	Ts2      int64  `json:"ts2"`
	Ts3      int64  `json:"ts3"`
}

type Request struct {
	Amount *int32 `json:"amount"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MessageType3 struct {
	Type  string `json:"type"`
	Event any    `json:"event,omitempty"`
	Ts3   int64  `json:"ts3"`
}

const (
	codeInvalidParameterError string = "InvalidParameterError"
)
