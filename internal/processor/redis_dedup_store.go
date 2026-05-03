package processor

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisSetClient 定义 Redis set 操作，便于单元测试替换。
type RedisSetClient interface {
	SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Close() error
}

// RedisDedupStore 使用 Redis set 保存跨运行去重状态。
type RedisDedupStore struct {
	client RedisSetClient
	key    string
}

// NewRedisDedupStore 创建 Redis 去重状态存储。
func NewRedisDedupStore(client RedisSetClient, key string) *RedisDedupStore {
	return &RedisDedupStore{client: client, key: key}
}

// NewRedisClient 创建 go-redis 客户端。
func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// Seen 判断指定 key 是否已经处理过。
func (s *RedisDedupStore) Seen(ctx context.Context, key string) (bool, error) {
	return s.client.SIsMember(ctx, s.key, key).Result()
}

// Mark 记录指定 key 已经处理过。
func (s *RedisDedupStore) Mark(ctx context.Context, key string) error {
	return s.client.SAdd(ctx, s.key, key).Err()
}
