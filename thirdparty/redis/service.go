package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	Client *redis.Client
}

func NewRedisService(ctx context.Context, address, password string, db int) *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       db,       // use default DB
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatal(err)
	}

	return &RedisService{Client: rdb}
}
