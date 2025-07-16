package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	Client *redis.Client
}

// NewRedisService creates a new instance of RedisService with the specified
// connection options.
//
// It establishes a connection to the Redis server located at the given address
// using the provided password and database index. The function will log a fatal
// error if the connection cannot be established.
//
// Parameters:
//
//	ctx - Context for managing request-scoped values, cancelation signals, and deadlines.
//	address - The address of the Redis server.
//	password - The password for authenticating with the Redis server.
//	db - The database index to select within the Redis server.
//
// Returns:
//
//	A pointer to a newly created RedisService with an active Redis client.
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
