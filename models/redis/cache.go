package redis

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// var redisRwMtx *sync.Mutex No need for function based application. Needed for global variables and palces where writing is frequent
// var ctx = context.Background() // Do not use global context. As each operation might requeire differnt context with timeout
// Try to use context from api request. Which is r.Context pass it into get and set

func CreateConnection() *redis.Client {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "welcome@123", // No password set
		DB:       0,             // Use default DB
		Protocol: 2,             // Connection protocol
		// PoolSize: 10,
	})
	err := redisClient.Ping(context.TODO()).Err()

	if err != nil {
		log.Fatal("error connecting to redis server ", err)
	} else {
		slog.Info("Connection to redis successful")
	}

	return redisClient
}

func CacheRead(ctx context.Context, key string) (string, error) { // Here in ctx try to pass api request r.Context here
	return redisClient.Get(ctx, key).Result()
}

func CacheWrite(ctx context.Context, key string, value string) error {
	return redisClient.Set(ctx, key, value, 0).Err()
}

func CacheWriteWithExpiry(ctx context.Context, key string, value any, expiry int) error {
	fmt.Println("this is expiry", expiry)
	ti := time.Duration(expiry) * time.Second
	return redisClient.Set(ctx, key, value, ti).Err()
}

func CacheKeyTTL(ctx context.Context, key string) (time.Duration, error) {
	return redisClient.TTL(ctx, key).Result()
}
