package cachedemo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	libstore "github.com/eko/gocache/lib/v4/store"
	"github.com/prismgo/framework/cache"
)

// exampleStore is a small custom backend whose close state makes ownership visible.
type exampleStore struct {
	mu     sync.Mutex
	data   map[string]any
	closed bool
}

func (s *exampleStore) Get(_ context.Context, key any) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[fmt.Sprint(key)]
	if !ok {
		return nil, libstore.NotFoundWithCause(cache.ErrCacheMiss)
	}
	return value, nil
}

func (s *exampleStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, 0, err
}

func (s *exampleStore) Set(_ context.Context, key any, value any, _ ...libstore.Option) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[fmt.Sprint(key)] = value
	return nil
}

func (s *exampleStore) Delete(_ context.Context, key any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, fmt.Sprint(key))
	return nil
}

func (s *exampleStore) Invalidate(context.Context, ...libstore.InvalidateOption) error { return nil }

func (s *exampleStore) Clear(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = map[string]any{}
	return nil
}

func (s *exampleStore) GetType() string { return "demo-custom" }

func (s *exampleStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *exampleStore) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func customManager() (*cache.Manager, *exampleStore, error) {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return nil, nil, fmt.Errorf("create custom driver name: %w", err)
	}
	driver := "demo-example-memory-" + hex.EncodeToString(suffix[:])
	store := &exampleStore{data: map[string]any{}}
	cache.Extend(driver, func(ctx cache.StoreFactoryContext) (cache.StoreDriver, error) {
		if ctx.Name != "example" || ctx.Prefix != "demo:tenant" || ctx.Config.Options["scope"] != "demo" {
			return cache.StoreDriver{}, fmt.Errorf("custom driver context = %#v, want example/demo:tenant/scope=demo", ctx)
		}
		return cache.NewStoreDriver(store), nil
	})
	m, err := cache.NewManager(cache.Config{
		Default: "example", Prefix: "demo", Lock: cache.LockConfig{Prefix: "locks"},
		Stores: map[string]cache.StoreConfig{"example": {
			Driver: driver, Prefix: "tenant", Options: map[string]any{"scope": "demo"},
		}},
	})
	return m, store, err
}

func customDriverExample(ctx context.Context, key string) (result string, details []string, resultErr error) {
	m, store, err := customManager()
	if err != nil {
		return "", nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, m.Close()) }()
	repo := m.Default()
	if err := repo.Put(ctx, key, "custom-value", time.Minute); err != nil {
		return "", nil, err
	}
	value, err := repo.Get(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if _, err := store.Get(ctx, "demo:tenant:"+key); err != nil {
		return "", nil, fmt.Errorf("read custom store prefixed key %q: %w", "demo:tenant:"+key, err)
	}
	return fmt.Sprintf("value=%v; prefix=%s", value, repo.GetStore().Prefix()), []string{"driver:demo-example-memory"}, nil
}

func lifecycleExample(ctx context.Context, key string) (string, []string, error) {
	m, store, err := customManager()
	if err != nil {
		return "", nil, err
	}
	if err := m.Default().Put(ctx, key, "owned", time.Minute); err != nil {
		return "", nil, errors.Join(err, m.Close())
	}
	if err := m.Close(); err != nil {
		return "", nil, err
	}
	if !store.isClosed() {
		return "", nil, fmt.Errorf("custom store closed = false, want true after manager.Close")
	}
	return "manager_closed=true", nil, nil
}
