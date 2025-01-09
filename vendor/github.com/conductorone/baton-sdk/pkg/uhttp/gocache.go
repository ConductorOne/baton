package uhttp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/maypok86/otter"
	"go.uber.org/zap"
)

const (
	cacheTTLMaximum  uint64 = 31536000 // 31536000 seconds = one year
	cacheTTLDefault  uint64 = 3600     // 3600 seconds = one hour
	defaultCacheSize uint   = 5        // MB
)

type CacheBackend string

const (
	CacheBackendDB     CacheBackend = "db"
	CacheBackendMemory CacheBackend = "memory"
	CacheBackendNoop   CacheBackend = "noop"
)

type CacheConfig struct {
	LogDebug bool
	TTL      uint64       // If 0, cache is disabled
	MaxSize  uint         // MB
	Backend  CacheBackend // If noop, cache is disabled
}

type CacheStats struct {
	Hits   int64
	Misses int64
}

type ContextKey struct{}

type GoCache struct {
	rootLibrary *otter.Cache[string, []byte]
}

type NoopCache struct {
	counter int64
}

func NewNoopCache(ctx context.Context) *NoopCache {
	return &NoopCache{}
}

func (g *NoopCache) Get(req *http.Request) (*http.Response, error) {
	// This isn't threadsafe but who cares? It's the noop cache.
	g.counter++
	return nil, nil
}

func (n *NoopCache) Set(req *http.Request, value *http.Response) error {
	return nil
}

func (n *NoopCache) Clear(ctx context.Context) error {
	return nil
}

func (n *NoopCache) Stats(ctx context.Context) CacheStats {
	return CacheStats{
		Hits:   0,
		Misses: n.counter,
	}
}

func (cc *CacheConfig) ToString() string {
	return fmt.Sprintf("Backend: %v, TTL: %d, MaxSize: %dMB, LogDebug: %t", cc.Backend, cc.TTL, cc.MaxSize, cc.LogDebug)
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:      cacheTTLDefault,
		MaxSize:  defaultCacheSize,
		LogDebug: false,
		Backend:  CacheBackendMemory,
	}
}

func NewCacheConfigFromEnv() *CacheConfig {
	config := DefaultCacheConfig()

	cacheMaxSize, err := strconv.ParseInt(os.Getenv("BATON_HTTP_CACHE_MAX_SIZE"), 10, 64)
	if err == nil && cacheMaxSize >= 0 {
		config.MaxSize = uint(cacheMaxSize)
	}

	cacheTTL, err := strconv.ParseUint(os.Getenv("BATON_HTTP_CACHE_TTL"), 10, 64)
	if err == nil {
		config.TTL = min(cacheTTLMaximum, max(0, cacheTTL))
	}

	cacheBackend := os.Getenv("BATON_HTTP_CACHE_BACKEND")
	switch cacheBackend {
	case "db":
		config.Backend = CacheBackendDB
	case "memory":
		config.Backend = CacheBackendMemory
	case "noop":
		config.Backend = CacheBackendNoop
	}

	disableCache, err := strconv.ParseBool(os.Getenv("BATON_DISABLE_HTTP_CACHE"))
	if err != nil {
		disableCache = false
	}
	if disableCache {
		config.Backend = CacheBackendNoop
	}

	return &config
}

func NewCacheConfigFromCtx(ctx context.Context) (*CacheConfig, error) {
	defaultConfig := DefaultCacheConfig()
	if v := ctx.Value(ContextKey{}); v != nil {
		ctxConfig, ok := v.(CacheConfig)
		if !ok {
			return nil, fmt.Errorf("error casting config values from context")
		}
		return &ctxConfig, nil
	}
	return &defaultConfig, nil
}

func NewHttpCache(ctx context.Context, config *CacheConfig) (icache, error) {
	l := ctxzap.Extract(ctx)

	if config == nil {
		config = NewCacheConfigFromEnv()
	}

	l.Info("http cache config", zap.String("config", config.ToString()))

	if config.TTL == 0 {
		l.Debug("CacheTTL is 0, disabling cache.", zap.Uint64("CacheTTL", config.TTL))
		return NewNoopCache(ctx), nil
	}

	switch config.Backend {
	case CacheBackendNoop:
		l.Debug("Using noop cache")
		return NewNoopCache(ctx), nil
	case CacheBackendMemory:
		l.Debug("Using in-memory cache")
		memCache, err := NewGoCache(ctx, *config)
		if err != nil {
			l.Error("error creating http cache (in-memory)", zap.Error(err), zap.Any("config", *config))
			return nil, err
		}
		return memCache, nil
	case CacheBackendDB:
		l.Debug("Using db cache")
		dbCache, err := NewDBCache(ctx, *config)
		if err != nil {
			l.Error("error creating http cache (db-cache)", zap.Error(err), zap.Any("config", *config))
			return nil, err
		}
		return dbCache, nil
	}

	return NewNoopCache(ctx), nil
}

func NewGoCache(ctx context.Context, cfg CacheConfig) (*GoCache, error) {
	l := ctxzap.Extract(ctx)
	gc := GoCache{}
	maxSize := cfg.MaxSize * 1024 * 1024
	if maxSize > math.MaxInt {
		return nil, fmt.Errorf("error converting max size to bytes")
	}
	//nolint:gosec // disable G115: we check the max size
	cache, err := otter.MustBuilder[string, []byte](int(maxSize)).
		CollectStats().
		Cost(func(key string, value []byte) uint32 {
			return uint32(len(key) + len(value))
		}).
		WithTTL(time.Duration(cfg.TTL) * time.Second).
		Build()

	if err != nil {
		l.Error("cache initialization error", zap.Error(err))
		return nil, err
	}

	l.Debug("otter cache initialized", zap.Int("capacity", cache.Capacity()))
	gc.rootLibrary = &cache

	return &gc, nil
}

func (g *GoCache) Stats(ctx context.Context) CacheStats {
	if g.rootLibrary == nil {
		return CacheStats{}
	}
	stats := g.rootLibrary.Stats()
	return CacheStats{
		Hits:   stats.Hits(),
		Misses: stats.Misses(),
	}
}

func (g *GoCache) Get(req *http.Request) (*http.Response, error) {
	if g.rootLibrary == nil {
		return nil, nil
	}

	key, err := CreateCacheKey(req)
	if err != nil {
		return nil, err
	}

	entry, ok := g.rootLibrary.Get(key)
	if !ok {
		return nil, nil
	}

	if len(entry) == 0 {
		return nil, nil
	}

	r := bufio.NewReader(bytes.NewReader(entry))
	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *GoCache) Set(req *http.Request, value *http.Response) error {
	if g.rootLibrary == nil {
		return nil
	}

	key, err := CreateCacheKey(req)
	if err != nil {
		return err
	}

	newValue, err := httputil.DumpResponse(value, true)
	if err != nil {
		return err
	}

	// Otter's cost function rejects large responses if there's not enough room
	// TODO: return some error or warning that we couldn't set?
	_ = g.rootLibrary.Set(key, newValue)

	return nil
}

func (g *GoCache) Delete(key string) error {
	if g.rootLibrary == nil {
		return nil
	}

	g.rootLibrary.Delete(key)

	return nil
}

func (g *GoCache) Clear(ctx context.Context) error {
	l := ctxzap.Extract(ctx)
	if g.rootLibrary == nil {
		l.Debug("clear: rootLibrary is nil")
		return nil
	}

	g.rootLibrary.Clear()

	l.Debug("reset cache")
	return nil
}

func (g *GoCache) Has(key string) bool {
	if g.rootLibrary == nil {
		return false
	}
	_, found := g.rootLibrary.Get(key)
	return found
}
