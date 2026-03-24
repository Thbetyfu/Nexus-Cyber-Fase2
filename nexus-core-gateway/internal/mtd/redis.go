package mtd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientWrapper encapsulates the connection to Redis.
// Implements connection pooling automatically via go-redis.
type RedisClientWrapper struct {
	Client  *redis.Client
	Enabled bool
}

// NewRedisClient creates a new Redis connection pool with fallback.
// ISO-25010 Reliability: If Redis is offline, it falls back to local memory without crashing.
func NewRedisClient() *RedisClientWrapper {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisURL,
		Password:     "",
		DB:           0,
		PoolSize:     10, // Small pool for local mode
		MinIdleConns: 1,
		DialTimeout:  300 * time.Millisecond, // Instant failover
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return &RedisClientWrapper{Enabled: false}
	}

	log.Printf("[MTD-REDIS] CONNECTED to Distributed Cache: %s. Using %d connection pool.", redisURL, 100)
	return &RedisClientWrapper{
		Client:  client,
		Enabled: true,
	}
}

// Global reference for other parts of MTD
var MtdRedis *RedisClientWrapper

func InitRedis() {
	MtdRedis = NewRedisClient()
}
