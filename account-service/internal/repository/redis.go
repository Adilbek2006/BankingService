package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(host, port string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to Redis!")
	return &RedisCache{client: client}, nil
}
func (r *RedisCache) SaveToCache(ctx context.Context, accountID string, data string) error {
	return r.client.Set(ctx, "account:"+accountID, data, 5*time.Minute).Err()
}

func (r *RedisCache) GetFromCache(ctx context.Context, accountID string) (string, error) {
	return r.client.Get(ctx, "account:"+accountID).Result()
}
func (r *RedisCache) DeleteFromCache(ctx context.Context, accountID string) error {
	return r.client.Del(ctx, "account:"+accountID).Err()
}
