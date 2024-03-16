package edge

import (
	"context"
	"database/sql"
	db "edge/db/sqlc"
	logutil "edge/util/log"
	"time"

	"github.com/google/uuid"
)

type CoinAcceptorEventCtrl struct {
	store        db.IStore
	iotContainer *IotContainer
}

func NewCoinAcceptorEvenCtrl(store db.IStore, ic *IotContainer) *CoinAcceptorEventCtrl {
	return &CoinAcceptorEventCtrl{store: store, iotContainer: ic}
}

func (c *CoinAcceptorEventCtrl) HandleCoinInserted(deviceID string, amount int32, ts int64) {
	arg := db.CreateRecordParams{
		ID:         uuid.New(),
		DeviceID:   deviceID,
		Type:       db.RecordTypeCoinAcceptorCoinInserted,
		Amount:     amount,
		IsUploaded: false,
		UploadedAt: sql.NullInt64{Valid: false},
		Ts:         ts,
	}

	if iot := c.iotContainer.Get(); iot != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := iot.AddCoinAcceptorCoinInsertedRecord(ctx, arg.ID, arg.DeviceID, arg.Amount, arg.Ts); err != nil {
			logutil.GetLogger().Errorf("upload coin acceptor coin inserted record error, err=%s, record_id=%s, device_id=%s, amount=%d, ts=%d",
				err, arg.ID, arg.DeviceID, arg.Amount, arg.Ts)
		} else {
			arg.IsUploaded = true
			arg.UploadedAt = sql.NullInt64{Valid: true, Int64: time.Now().UnixMilli()}
		}
	}
	if _, err := c.store.CreateRecord(context.Background(), arg); err != nil {
		logutil.GetLogger().Errorf("create record error, err=%s, arg=%#v", err, arg)
	}
}

func (c *CoinAcceptorEventCtrl) HandleDeviceStatusChanged(deviceID string, status CoinAcceptorStatus) {
	if iot := c.iotContainer.Get(); iot != nil {
		iot.SendCoinAcceptorStatusChangedEvent(deviceID, status)
	}
}
