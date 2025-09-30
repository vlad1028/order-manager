package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/vlad1028/order-manager/internal/models/order"
	"log"
	"os"
	"time"
)

func MustNew(ctx context.Context, ttl time.Duration) *Redis {
	url := os.Getenv("REDIS_URL")
	pwd := os.Getenv("REDIS_PWD")

	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: pwd,
		DB:       0,
	})

	status := client.Ping(ctx)
	if status.Err() != nil {
		log.Fatalf("failed to connect to redis: %v", status.Err())
	}

	return &Redis{
		ttl:    ttl,
		client: client,
	}
}

// Redis Использую TTL как механизм инвалидации кэша.
// При этом данные могут быть устаревшими, если другие серверы обновили данные в бд.
// Но при небольших ttl это не критично.
type Redis struct {
	ttl    time.Duration
	client *redis.Client
}

func (r *Redis) Get(ctx context.Context, key string) (*order.Order, bool) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false
		}

		log.Printf("failed to fetch key %s: %v", key, err)
		return nil, false
	}

	var result *order.Order
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		log.Printf("failed to unmarshal key %s: %v", key, err)
		return nil, false
	}

	return result, true
}

func (r *Redis) Set(ctx context.Context, key string, order *order.Order) error {
	b, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	err = r.client.Set(ctx, key, b, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed write value to redis %w", err)
	}
	return nil
}
