package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/pagespeedonline/v5"
)

// CacheKey uniquely identifies a PageSpeed request configuration
type CacheKey struct {
	URL        string
	Strategy   string   // "mobile" or "desktop"
	Categories []string // sorted list of categories
	Locale     string
	Campaign   string
	Source     string
}

// String generates a unique string representation of the cache key
func (k CacheKey) String() string {
	// Sort categories to ensure consistent key generation
	sortedCategories := make([]string, len(k.Categories))
	copy(sortedCategories, k.Categories)
	sort.Strings(sortedCategories)

	// Create a deterministic string representation
	parts := []string{
		k.URL,
		k.Strategy,
		strings.Join(sortedCategories, ","),
		k.Locale,
		k.Campaign,
		k.Source,
	}

	// Generate SHA256 hash for shorter, fixed-length keys
	h := sha256.New()
	h.Write([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h.Sum(nil))
}

// CacheEntry represents a cached PageSpeed result
type CacheEntry struct {
	Key       CacheKey
	Result    *pagespeedonline.PagespeedApiPagespeedResponseV5
	Timestamp time.Time
	TTL       time.Duration
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Since(e.Timestamp) > e.TTL
}

// ResultCache defines the interface for caching PageSpeed results
type ResultCache interface {
	Get(key CacheKey) (*pagespeedonline.PagespeedApiPagespeedResponseV5, bool)
	Set(key CacheKey, result *pagespeedonline.PagespeedApiPagespeedResponseV5)
	Clear()
}

// InMemoryCache implements ResultCache using an in-memory map
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// New creates a new InMemoryCache with the specified TTL
func New(ttl time.Duration) *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached result if it exists and is not expired
func (c *InMemoryCache) Get(key CacheKey) (*pagespeedonline.PagespeedApiPagespeedResponseV5, bool) {
	if c.ttl == 0 {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	keyStr := key.String()
	entry, exists := c.entries[keyStr]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if entry.IsExpired() {
		// Use deferred cleanup to avoid holding write lock during read
		go func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			// Double-check the entry still exists and is expired
			if e, ok := c.entries[keyStr]; ok && e.IsExpired() {
				delete(c.entries, keyStr)
			}
		}()
		return nil, false
	}

	return entry.Result, true
}

// Set stores a result in the cache
func (c *InMemoryCache) Set(key CacheKey, result *pagespeedonline.PagespeedApiPagespeedResponseV5) {
	if c.ttl == 0 || result == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key.String()] = &CacheEntry{
		Key:       key,
		Result:    result,
		Timestamp: time.Now(),
		TTL:       c.ttl,
	}
}

// Clear removes all entries from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Size returns the current number of entries in the cache
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// NewCacheKey creates a CacheKey from request parameters
func NewCacheKey(url, strategy string, categories []string, locale, campaign, source string) CacheKey {
	// Make a copy of categories to avoid external modifications
	categoriesCopy := make([]string, len(categories))
	copy(categoriesCopy, categories)

	return CacheKey{
		URL:        url,
		Strategy:   strategy,
		Categories: categoriesCopy,
		Locale:     locale,
		Campaign:   campaign,
		Source:     source,
	}
}