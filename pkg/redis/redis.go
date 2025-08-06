package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"auth-service/internal/config"
)

// Client Redis 客户端包装器
type Client struct {
	rdb    *redis.Client
	prefix string
}

// NewClient 创建 Redis 客户端
func NewClient(cfg *config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Client{
		rdb:    rdb,
		prefix: cfg.Prefix,
	}
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Set 设置键值对
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := c.prefix + key
	return c.rdb.Set(ctx, fullKey, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	fullKey := c.prefix + key
	return c.rdb.Get(ctx, fullKey).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.prefix + key
	}
	return c.rdb.Del(ctx, fullKeys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.prefix + key
	result, err := c.rdb.Exists(ctx, fullKey).Result()
	return result > 0, err
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.rdb.Close()
}
