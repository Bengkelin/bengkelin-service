package redis

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

var (
    redisClient *redis.Client
    ctx         = context.Background()
)

type CacheInterface interface {
    Set(key string, value interface{}, expiration time.Duration) error
    Get(key string, dest interface{}) error
    Delete(key string) error
}

type RedisCache struct {
    client *redis.Client
}

func SetupRedis(addr, password string, db int) {
    redisClient = redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
}

func GetRedisClient() *RedisCache {
    return &RedisCache{client: redisClient}
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
    json, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, key, json, expiration).Err()
}

func (r *RedisCache) Get(key string, dest interface{}) error {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) Delete(key string) error {
    return r.client.Del(ctx, key).Err()
}