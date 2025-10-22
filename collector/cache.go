package collector

import (
	"sync"
	"time"
	"crypto/sha256"
	"encoding/json"
)

type cacheEntry struct {
	Result    *ScrapeResult
	ExpiresAt time.Time
}

type scrapeCache struct {
	entries map[string]cacheEntry
	mutex   sync.Mutex
	ttl     time.Duration
}

func newScrapeCache(ttl time.Duration) *scrapeCache {
	if ttl <= 0 {
		return nil
	}
	return &scrapeCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *scrapeCache) get(key string) (*ScrapeResult, bool) {
	if c == nil {
		return nil, false
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		if ok {
			delete(c.entries, key)
		}
		return nil, false
	}
	return entry.Result, true
}

func (c *scrapeCache) set(key string, result *ScrapeResult) {
	if c == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[key] = cacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func cacheKeyFromRequest(req ScrapeRequest) string {
	b, _ := json.Marshal(req)
	h := sha256.Sum256(b)
	return string(h[:])
}
