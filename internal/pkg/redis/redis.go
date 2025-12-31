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
        Addr:         addr,
        Password:     password,
        DB:           db,
        DialTimeout:  10 * time.Second,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 5 * time.Second,
    })
}

func SetupRedisFromURL(redisURL string) {
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        panic("Failed to parse Redis URL: " + err.Error())
    }
    
    // Set connection timeouts
    opt.DialTimeout = 10 * time.Second
    opt.ReadTimeout = 5 * time.Second
    opt.WriteTimeout = 5 * time.Second
    
    redisClient = redis.NewClient(opt)
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

func (r *RedisCache) SetWithContext(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
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

func (r *RedisCache) GetWithContext(ctx context.Context, key string, dest interface{}) error {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) Delete(key string) error {
    return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) DeleteWithContext(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}
// GetClient returns the underlying Redis client for advanced operations
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}