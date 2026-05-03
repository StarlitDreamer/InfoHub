package processor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRedisDedupStoreIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip redis integration test in short mode")
	}

	addr := strings.TrimSpace(os.Getenv("INFOHUB_TEST_REDIS_ADDR"))
	if addr == "" {
		t.Skip("skip redis integration test without INFOHUB_TEST_REDIS_ADDR")
	}

	password := os.Getenv("INFOHUB_TEST_REDIS_PASSWORD")
	db := 0

	client := NewRedisClient(addr, password, db)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping redis failed: %v", err)
	}

	key := fmt.Sprintf("infohub:dedup:test:%d", time.Now().UnixNano())
	store := NewRedisDedupStore(client, key)
	defer client.Del(context.Background(), key)

	seen, err := store.Seen(ctx, "item-1")
	if err != nil {
		t.Fatalf("check redis dedup status failed: %v", err)
	}
	if seen {
		t.Fatal("expected item-1 to be unseen before mark")
	}

	if err := store.Mark(ctx, "item-1"); err != nil {
		t.Fatalf("mark redis dedup status failed: %v", err)
	}

	seen, err = store.Seen(ctx, "item-1")
	if err != nil {
		t.Fatalf("recheck redis dedup status failed: %v", err)
	}
	if !seen {
		t.Fatal("expected item-1 to be seen after mark")
	}

	if err := store.Mark(ctx, "item-1"); err != nil {
		t.Fatalf("mark duplicate redis dedup status failed: %v", err)
	}

	size, err := client.SCard(ctx, key).Result()
	if err != nil {
		t.Fatalf("read redis set size failed: %v", err)
	}
	if size != 1 {
		t.Fatalf("expected redis set size 1 after duplicate mark, got %d", size)
	}
}
