package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/redis/go-redis/v9"
)

type Cache struct{ Client *redis.Client }

func Open(ctx context.Context, cfg config.RedisConfig) (*Cache, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	client := redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Cache{Client: client}, nil
}
func (c *Cache) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Close()
}
func (c *Cache) GetJSON(ctx context.Context, key string, dest any) error {
	if c == nil {
		return redis.Nil
	}
	value, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}
func (c *Cache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, key, data, ttl).Err()
}
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if c == nil || len(keys) == 0 {
		return nil
	}
	return c.Client.Del(ctx, keys...).Err()
}

func (c *Cache) DeletePrefix(ctx context.Context, prefix string) error {
	if c == nil {
		return nil
	}
	var keys []string
	iterator := c.Client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iterator.Next(ctx) {
		keys = append(keys, iterator.Val())
	}
	if err := iterator.Err(); err != nil {
		return err
	}
	return c.Delete(ctx, keys...)
}
