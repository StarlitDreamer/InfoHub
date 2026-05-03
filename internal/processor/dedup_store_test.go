package processor

import (
	"context"
	"path/filepath"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestFileDedupStorePersistsKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seen.json")
	store := NewFileDedupStore(path)
	key := DedupKey(model.NewsItem{URL: "https://example.com/a"})

	seen, err := store.Seen(context.Background(), key)
	if err != nil {
		t.Fatalf("read dedup status failed: %v", err)
	}
	if seen {
		t.Fatal("expected new key to be unseen")
	}

	if err := store.Mark(context.Background(), key); err != nil {
		t.Fatalf("write dedup status failed: %v", err)
	}

	nextStore := NewFileDedupStore(path)
	seen, err = nextStore.Seen(context.Background(), key)
	if err != nil {
		t.Fatalf("reload dedup status failed: %v", err)
	}
	if !seen {
		t.Fatal("expected persisted key to be visible across store instances")
	}
}

func TestDedupKeysIncludeTitleURLAndContentFingerprints(t *testing.T) {
	keys := DedupKeys(model.NewsItem{
		Title:   "Same Title",
		URL:     "https://example.com/post/",
		Content: "Same   body text",
	})

	if len(keys) != 3 {
		t.Fatalf("expected 3 dedup keys, got %d", len(keys))
	}
	if keys[0] == "" || keys[1] == "" || keys[2] == "" {
		t.Fatalf("expected non-empty dedup keys, got %+v", keys)
	}
}

func TestDedupKeyPrefersURLFingerprint(t *testing.T) {
	key := DedupKey(model.NewsItem{
		Title:   "Same Title",
		URL:     "https://example.com/post/",
		Content: "Same body text",
	})

	if len(key) == 0 || key[:4] != "url:" {
		t.Fatalf("expected single dedup key to prefer url fingerprint, got %s", key)
	}
}

func TestDedupKeysNormalizeEquivalentValues(t *testing.T) {
	left := DedupKeys(model.NewsItem{
		Title:   "Same Title",
		URL:     "https://example.com/post/",
		Content: "Same   body text",
	})
	right := DedupKeys(model.NewsItem{
		Title:   " same   title ",
		URL:     "https://example.com/post",
		Content: " same body text ",
	})

	if len(left) != len(right) {
		t.Fatalf("expected equivalent items to generate same key count, got %d vs %d", len(left), len(right))
	}
	for index := range left {
		if left[index] != right[index] {
			t.Fatalf("expected equivalent items to generate same dedup keys, got %+v vs %+v", left, right)
		}
	}
}
