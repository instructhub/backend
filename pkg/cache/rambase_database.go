package cache

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	RedisLimiter *redis.Client
)

// InitializeRedis initializes the Redis client.
func init() {
	// Read Redis database number.
	limiterDB := 0
	normalDB := 1

	host := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	if host == "" || port == "" {
		log.Fatalf("CACHE_HOST or CACHE_PORT is not set")
	}

	url := fmt.Sprintf("%s:%s", host, port)

	// Set Redis options.
	options := &redis.Options{
		Addr:         url,
		Password:     os.Getenv("CACHE_PASSWORD"),
		DB:           normalDB,
		PoolSize:     10,
		MinIdleConns: 1,
	}

	RedisClient = redis.NewClient(options)

	options = &redis.Options{
		Addr:         url,
		Password:     os.Getenv("CACHE_PASSWORD"),
		DB:           limiterDB,
		PoolSize:     10,
		MinIdleConns: 1,
	}
	
	RedisLimiter = redis.NewClient(options)

	// Test connection
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}

	if err := RedisLimiter.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}
}
