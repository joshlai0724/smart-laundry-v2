package heartbeat

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func SendHeartbeat(client *redis.Client, key string, expiration time.Duration) error {
	ctx := context.Background()
	_, err := client.Set(ctx, key, time.Now().UnixMilli(), expiration).Result()
	return err
}

func StopHeartbeat(client *redis.Client, key string) error {
	ctx := context.Background()
	_, err := client.Del(ctx, key).Result()
	return err
}

func CheckHeartbeat(client *redis.Client, key string) (bool, error) {
	ctx := context.Background()
	n, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		return true, nil
	}
	return false, nil
}

var prefix = "heartbeat:"

func GetStoreIDHeartbeatName(storeID string) string {
	return prefix + "store-id:" + storeID
}
