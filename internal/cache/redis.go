package cache

import (
	"context"
	"fmt"
	"time"

	"go-admin/config"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init() error {
	cfg := config.Cfg.Redis
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}
	return nil
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

func Del(ctx context.Context, keys ...string) error {
	return RDB.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, keys ...string) (bool, error) {
	n, err := RDB.Exists(ctx, keys...).Result()
	return n > 0, err
}

func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return RDB.SetNX(ctx, key, value, expiration).Result()
}

func Incr(ctx context.Context, key string) (int64, error) {
	return RDB.Incr(ctx, key).Result()
}

func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return RDB.Expire(ctx, key, expiration).Err()
}

// Token 黑名单：吊销 access token
func RevokeToken(ctx context.Context, token string, expiration time.Duration) error {
	return RDB.Set(ctx, "token:blacklist:"+token, "1", expiration).Err()
}

func IsTokenRevoked(ctx context.Context, token string) bool {
	n, _ := RDB.Exists(ctx, "token:blacklist:"+token).Result()
	return n > 0
}
