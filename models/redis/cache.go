package redis

import (
	"context"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisRwMtx *sync.Mutex
var ctx = context.Background()

func CreateConnection() *redis.Client {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "welcome@123", // No password set
		DB:       0,             // Use default DB
		Protocol: 2,             // Connection protocol
	})
	err := redisClient.Ping(ctx).Err()

	if err != nil {
		log.Fatal("error connecting to redis server ", err)
	} else {
		slog.Info("Connection to redis successful")
	}

	return redisClient
}

func CacheRead() {

}

func CacheWrite() {

}

func CacheWriteWithTime() {

}

func KeyTimeInterval() {

}
