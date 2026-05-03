package processor

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisDedupStore(t *testing.T) {
	client := newFakeRedisSetClient()
	store := NewRedisDedupStore(client, "infohub:dedup:seen")
	ctx := context.Background()

	seen, err := store.Seen(ctx, "key-1")
	if err != nil {
		t.Fatalf("读取 Redis 去重状态失败：%v", err)
	}

	if seen {
		t.Fatal("新 key 不应被判断为已处理")
	}

	if err := store.Mark(ctx, "key-1"); err != nil {
		t.Fatalf("写入 Redis 去重状态失败：%v", err)
	}

	seen, err = store.Seen(ctx, "key-1")
	if err != nil {
		t.Fatalf("重新读取 Redis 去重状态失败：%v", err)
	}

	if !seen {
		t.Fatal("期望识别已处理 key")
	}
}

type fakeRedisSetClient struct {
	values map[string]map[string]struct{}
}

func newFakeRedisSetClient() *fakeRedisSetClient {
	return &fakeRedisSetClient{values: map[string]map[string]struct{}{}}
}

func (c *fakeRedisSetClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	value, ok := member.(string)
	if !ok {
		cmd.SetVal(false)
		return cmd
	}

	_, exists := c.values[key][value]
	cmd.SetVal(exists)
	return cmd
}

func (c *fakeRedisSetClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	if _, ok := c.values[key]; !ok {
		c.values[key] = map[string]struct{}{}
	}

	var added int64
	for _, member := range members {
		value, ok := member.(string)
		if !ok {
			continue
		}

		if _, exists := c.values[key][value]; !exists {
			c.values[key][value] = struct{}{}
			added++
		}
	}

	cmd.SetVal(added)
	return cmd
}

func (c *fakeRedisSetClient) Close() error {
	return nil
}
