package redis

import (
	"context"
	"log"
	"queue-system/config"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis(cfg *config.Config) *redis.Client{
	RDB = redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB: 0,
	})

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	log.Println("Connected to Redis successfully")
	return RDB
}