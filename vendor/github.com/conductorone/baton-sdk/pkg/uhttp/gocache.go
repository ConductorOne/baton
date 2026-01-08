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
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
	"go.uber.org/zap"
)

const (
	cacheTTLMaximum    time.Duration = 31536000 * time.Second // 31536000 seconds = one year
	cacheTTLDefault    time.Duration = 3600 * time.Second     // 3600 seconds = one hour
	defaultCacheSizeMb uint64        = 5                      // MB
)

type CacheBackend string

const (
	CacheBackendDB     CacheBackend = "db"
	CacheBackendMemory CacheBackend = "memory"
	CacheBackendNoop   CacheBackend = "noop"
)

type CacheConfig struct {
	LogDebug  bool
	TTL       time.Duration // If 0, cache is disabled
	MaxSizeMb uint64        // MB
	Backend   CacheBackend  // If noop, cache is disabled
}

type CacheStats struct {
	Hits   uint64
	Misses uint64
}

type ContextKey struct{}

type GoCache struct {
	rootLibrary *otter.Cache[string, []byte]
}

type NoopCache struct {
	counter uint64
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
	return fmt.Sprintf("Backend: %v, TTL: %d, MaxSize: %dMB, LogDebug: %t", cc.Backend, cc.TTL, cc.MaxSizeMb, cc.LogDebug)
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:       cacheTTLDefault,
		MaxSizeMb: defaultCacheSizeMb,
		LogDebug:  false,
		Backend:   CacheBackendMemory,
	}
}

func NewCacheConfigFromEnv() *CacheConfig {
	config := DefaultCacheConfig()

	cacheMaxSize, err := strconv.ParseInt(os.Getenv("BATON_HTTP_CACHE_MAX_SIZE"), 10, 64)
	if err == nil && cacheMaxSize >= 0 {
		config.MaxSizeMb = uint64(cacheMaxSize)
	}

	cacheTTL, err := strconv.ParseInt(os.Getenv("BATON_HTTP_CACHE_TTL"), 10, 64)
	if err == nil {
		config.TTL = min(cacheTTLMaximum, max(0, time.Duration(cacheTTL)*time.Second))
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
		l.Debug("NewHttpCache: Cache TTL is 0, disabling cache.", zap.Duration("cache_ttl", config.TTL))
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
	maxSize := cfg.MaxSizeMb * 1024 * 1024
	if maxSize > math.MaxInt {
		return nil, fmt.Errorf("error converting max size to bytes")
	}
	cache, err := otter.New(&otter.Options[string, []byte]{
		MaximumWeight: maxSize,
		StatsRecorder: stats.NewCounter(),
		Weigher: func(key string, value []byte) uint32 {
			weight64 := uint64(len(key)) + uint64(len(value))
			if weight64 > uint64(math.MaxUint32) {
				return math.MaxUint32
			}
			return uint32(weight64)
		},
		ExpiryCalculator: otter.ExpiryWriting[string, []byte](cfg.TTL),
	})

	if err != nil {
		l.Error("cache initialization error", zap.Error(err))
		return nil, err
	}

	l.Debug("otter cache initialized", zap.Uint64("capacity", cache.GetMaximum()))
	gc.rootLibrary = cache

	return &gc, nil
}

func (g *GoCache) Stats(ctx context.Context) CacheStats {
	if g.rootLibrary == nil {
		return CacheStats{}
	}
	stats := g.rootLibrary.Stats()
	return CacheStats{
		Hits:   stats.Hits,
		Misses: stats.Misses,
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

	entry, found := g.rootLibrary.GetIfPresent(key)
	if !found {
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
	_, _ = g.rootLibrary.Set(key, newValue)

	return nil
}

func (g *GoCache) Delete(key string) error {
	if g.rootLibrary == nil {
		return nil
	}

	g.rootLibrary.Invalidate(key)

	return nil
}

func (g *GoCache) Clear(ctx context.Context) error {
	l := ctxzap.Extract(ctx)
	if g.rootLibrary == nil {
		l.Debug("clear: rootLibrary is nil")
		return nil
	}

	g.rootLibrary.InvalidateAll()

	l.Debug("reset cache")
	return nil
}

func (g *GoCache) Has(key string) bool {
	if g.rootLibrary == nil {
		return false
	}
	_, found := g.rootLibrary.GetIfPresent(key)
	return found
}
