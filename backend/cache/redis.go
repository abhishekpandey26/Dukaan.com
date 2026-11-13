// Package cache is a tiny wrapper around Redis that stores Go values as JSON.
//
// Design choice: the cache is OPTIONAL. If Redis is down, Get just reports a
// "miss" and the app reads from Postgres instead. A cache should make things
// faster, never make them break.
package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client // nil means caching is disabled
}

// New connects to Redis. Pass an empty addr to disable caching completely.
func New(addr string) *Cache {
	if addr == "" {
		log.Println("REDIS_ADDR is empty - caching disabled")
		return &Cache{}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  500 * time.Millisecond, // fail fast so a dead Redis doesn't slow requests
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: redis at %s not reachable (%v) - will keep trying, requests fall back to postgres", addr, err)
	} else {
		log.Printf("connected to redis at %s", addr)
	}
	return &Cache{client: client}
}

// Get loads key into dest. It returns true only on a cache hit.
func (c *Cache) Get(ctx context.Context, key string, dest any) bool {
	if c.client == nil {
		return false
	}
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil { // redis.Nil (key not found) or a connection error -> treat as miss
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

// Set stores value as JSON for the given time-to-live.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	if c.client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	c.client.Set(ctx, key, data, ttl)
}

// DeletePattern removes all keys matching a pattern like "products:*".
// We use SCAN (not KEYS) because KEYS blocks Redis on large databases.
func (c *Cache) DeletePattern(ctx context.Context, pattern string) {
	if c.client == nil {
		return
	}
	var keys []string
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		c.client.Del(ctx, keys...)
	}
}

func (c *Cache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
