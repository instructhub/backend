package cache

import (
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// RedisConnection func for connect to Redis server.
func CacheConnection() (*redis.Client, error) {
	// Define Redis database number.
	dbNumber, _ := strconv.Atoi(os.Getenv("REDIS_DB_NUMBER"))

	url := fmt.Sprintf(
		"%s:%s",
		os.Getenv("CACHE_HOST"),
		os.Getenv("CACHE_PORT"),
	)

	// Set Redis options.
	options := &redis.Options{
		Addr:     url,
		Password: os.Getenv("CACHE_PASSWORD"),
		DB:       dbNumber,
		PoolSize:     10,
		MinIdleConns: 2,
	}

	return redis.NewClient(options), nil
}
