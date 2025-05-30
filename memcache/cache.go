package memcache

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/flemeur/taleplade"
)

const ErrCacheMiss = taleplade.Error("memcache: cache miss")

type Cache struct {
	mc *memcache.Client
}

func New(server ...string) *Cache {
	return &Cache{
		mc: memcache.New(server...),
	}
}

func (c *Cache) Ping() error {
	return c.mc.Ping()
}

func (c *Cache) Get(k string) ([]byte, error) {
	item, err := c.mc.Get(k)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	return item.Value, nil
}

func (c *Cache) Set(k string, b []byte) error {
	return c.mc.Set(&memcache.Item{Key: k, Value: b})
}
