package memcached

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/logger"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache"
)

var once sync.Once

var client *otelmemcache.Client

var memoryCache = cache.New(time.Minute*5, time.Minute*10)

// InitCache will create and load the cache. It is safe to call more than once
func InitCache(c *config.Config) {
	once.Do(func() {
		c := memcache.New(c.GetString("memcache.url"))
		client = otelmemcache.NewClientWithTracing(c)
	})
}

// GetCache will get the created cache client. Please call init cache before this
func GetCache() *otelmemcache.Client {
	return client
}

// GetItem will get an item from the cache and will fail silently if it does fail
func GetItem(ctx context.Context, m *otelmemcache.Client, key string, dest interface{}, useInMemory bool) bool {
	var data []byte
	out, ok := memoryCache.Get(key)
	if !ok || !useInMemory {
		item, err := m.WithContext(ctx).Get(key)
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false
		}
		if err != nil {
			logger.GetLogger(ctx).ErrorCtx(ctx, "unable to get cache item", logger.ErrField(err), logger.Any("key", key))
			return false
		}
		data = item.Value
	} else {
		data = out.([]byte)
	}

	if err := json.Unmarshal(data, &dest); err != nil {
		logger.GetLogger(ctx).ErrorCtx(ctx, "unable to unmarshal cached item", logger.ErrField(err), logger.Any("key", key))
		return false
	}

	return true
}

type defaults struct {
	Expiration int
}

// OneMinute is one minute in duration
const OneMinute = 60

// OneHour will be on hour till expiration
const OneHour = OneMinute * 60

// OneDay will be one day till expiration
const OneDay = OneHour * 24

// ThirtyDays will be thirty days till expiration
const ThirtyDays = OneDay * 30

// WithExpiration will set the item with an option to expire. This is in seconds
func WithExpiration(e int) func(d *defaults) {
	return func(d *defaults) {
		d.Expiration = e
	}
}

// SetItem will set an item in our cache and it will fail silently
func SetItem(ctx context.Context, m *otelmemcache.Client, key string, item interface{}, opts ...func(*defaults)) {
	d := defaults{}
	for _, o := range opts {
		o(&d)
	}

	b, err := json.Marshal(item)
	if err != nil {
		logger.GetLogger(ctx).ErrorCtx(ctx, "unable to marshal item", logger.ErrField(err), logger.Any("key", key))
		return
	}

	if err := m.WithContext(ctx).Set(&memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: int32(d.Expiration),
	}); err != nil {
		logger.GetLogger(ctx).ErrorCtx(ctx, "unable to set item", logger.ErrField(err), logger.Any("key", key))
	}
	setMemCache(key, b, d.Expiration)
}

func setMemCache(key string, item []byte, duration int) {
	dur := time.Duration(duration) * time.Second
	if dur == 0 {
		dur = -1
	}
	memoryCache.Set(key, item, dur)
}
