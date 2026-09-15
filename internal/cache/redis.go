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

// ---- refresh token 相关键 ----
//
// 每个 refresh token 自身是一个键（存储 userID），
// 另外用「用户 -> token 集合」维护索引，用于改密/禁用/登出时批量吊销。
//
// 早前的实现直接删 refresh_token:user:<id>，但这个键从未被写入过 ——
// 写入时用的是 refresh_token:<随机串>。键名不匹配导致吊销静默失效，
// 改密或禁用后旧 refresh token 依然能换发新的 access token。
func RefreshTokenKey(token string) string {
	return "refresh_token:" + token
}

// RefreshTokenSetKey 用户当前有效的 refresh token 集合
func RefreshTokenSetKey(userID uint) string {
	return fmt.Sprintf("refresh_token:user:%d", userID)
}

func SAdd(ctx context.Context, key string, members ...interface{}) error {
	return RDB.SAdd(ctx, key, members...).Err()
}

func SRem(ctx context.Context, key string, members ...interface{}) error {
	return RDB.SRem(ctx, key, members...).Err()
}

func SMembers(ctx context.Context, key string) ([]string, error) {
	return RDB.SMembers(ctx, key).Result()
}

// Token 黑名单：吊销 access token
func RevokeToken(ctx context.Context, token string, expiration time.Duration) error {
	return RDB.Set(ctx, "token:blacklist:"+token, "1", expiration).Err()
}

func IsTokenRevoked(ctx context.Context, token string) bool {
	n, _ := RDB.Exists(ctx, "token:blacklist:"+token).Result()
	return n > 0
}

// DelByPrefix 按前缀批量删除键。
// 使用 SCAN 分批遍历，避免 KEYS 命令在大 key 空间下阻塞 Redis。
func DelByPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		keys, next, err := RDB.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := RDB.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}
