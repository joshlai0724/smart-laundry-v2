package web

import (
	db "backend/db/sqlc"
	"time"
)

func isVerCodeValid(code db.VerCode) bool {
	if code.IsBlocked {
		return false
	}
	if time.Now().After(time.UnixMilli(code.ExpiredAt)) {
		return false
	}
	return true
}

func dtoRecordType2DbRecordType(_type string) []string {
	types := make([]string, 0)

	if _type == "all" || _type == "top-up" {
		types = append(types,
			db.RecordTypeCashTopUp,
		)
	}
	if _type == "all" || _type == "device" {
		types = append(types,
			db.RecordTypeCoinAcceptorRemoteInsertCoins,
		)
	}
	return types
}
