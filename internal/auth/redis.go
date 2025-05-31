package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenStore struct {
	Client *redis.Client
	TTL    time.Duration
}

func NewRedisTokenStore(addr string, db int, ttl time.Duration) *RedisTokenStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	return &RedisTokenStore{Client: client, TTL: ttl}
}

func (r *RedisTokenStore) SaveToken(ctx context.Context, token string, userID uint) error {
	return r.Client.Set(ctx, token, userID, r.TTL).Err()
}

func (r *RedisTokenStore) GetUserID(ctx context.Context, token string) (uint, error) {
	val, err := r.Client.Get(ctx, token).Result()
	if err != nil {
		return 0, err
	}
	var userID uint
	_, err = fmt.Sscanf(val, "%d", &userID)
	return userID, err
}

func (r *RedisTokenStore) RefreshToken(ctx context.Context, token string) error {
	return r.Client.Expire(ctx, token, r.TTL).Err()
}
