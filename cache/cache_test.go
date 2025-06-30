package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/pagespeedonline/v5"
)

func TestCacheKey_String(t *testing.T) {
	tests := []struct {
		name     string
		key1     CacheKey
		key2     CacheKey
		wantSame bool
	}{
		{
			name: "identical keys produce same string",
			key1: CacheKey{
				URL:        "https://example.com",
				Strategy:   "mobile",
				Categories: []string{"performance", "seo"},
				Locale:     "en",
				Campaign:   "test",
				Source:     "source1",
			},
			key2: CacheKey{
				URL:        "https://example.com",
				Strategy:   "mobile",
				Categories: []string{"performance", "seo"},
				Locale:     "en",
				Campaign:   "test",
				Source:     "source1",
			},
			wantSame: true,
		},
		{
			name: "different URL produces different string",
			key1: CacheKey{
				URL:        "https://example.com",
				Strategy:   "mobile",
				Categories: []string{"performance"},
			},
			key2: CacheKey{
				URL:        "https://different.com",
				Strategy:   "mobile",
				Categories: []string{"performance"},
			},
			wantSame: false,
		},
		{
			name: "different order of categories produces same string",
			key1: CacheKey{
				URL:        "https://example.com",
				Strategy:   "mobile",
				Categories: []string{"seo", "performance"},
			},
			key2: CacheKey{
				URL:        "https://example.com",
				Strategy:   "mobile",
				Categories: []string{"performance", "seo"},
			},
			wantSame: true,
		},
		{
			name: "different strategy produces different string",
			key1: CacheKey{
				URL:        "https://example.com",
				Strategy:   "mobile",
				Categories: []string{"performance"},
			},
			key2: CacheKey{
				URL:        "https://example.com",
				Strategy:   "desktop",
				Categories: []string{"performance"},
			},
			wantSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str1 := tt.key1.String()
			str2 := tt.key2.String()

			if tt.wantSame {
				assert.Equal(t, str1, str2)
			} else {
				assert.NotEqual(t, str1, str2)
			}
		})
	}
}

func TestCacheEntry_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		entry      CacheEntry
		wantExpired bool
	}{
		{
			name: "entry not expired",
			entry: CacheEntry{
				Timestamp: time.Now(),
				TTL:       1 * time.Hour,
			},
			wantExpired: false,
		},
		{
			name: "entry expired",
			entry: CacheEntry{
				Timestamp: time.Now().Add(-2 * time.Hour),
				TTL:       1 * time.Hour,
			},
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantExpired, tt.entry.IsExpired())
		})
	}
}

func TestInMemoryCache_GetSet(t *testing.T) {
	cache := New(5 * time.Minute)
	
	key := CacheKey{
		URL:        "https://example.com",
		Strategy:   "mobile",
		Categories: []string{"performance"},
	}
	
	result := &pagespeedonline.PagespeedApiPagespeedResponseV5{
		Id: "test-id",
	}

	// Test cache miss
	cachedResult, found := cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, cachedResult)

	// Test cache set and hit
	cache.Set(key, result)
	cachedResult, found = cache.Get(key)
	assert.True(t, found)
	assert.NotNil(t, cachedResult)
	assert.Equal(t, "test-id", cachedResult.Id)

	// Test cache size
	assert.Equal(t, 1, cache.Size())
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := New(100 * time.Millisecond)
	
	key := CacheKey{
		URL:      "https://example.com",
		Strategy: "mobile",
	}
	
	result := &pagespeedonline.PagespeedApiPagespeedResponseV5{
		Id: "test-id",
	}

	// Set and verify entry exists
	cache.Set(key, result)
	cachedResult, found := cache.Get(key)
	assert.True(t, found)
	assert.NotNil(t, cachedResult)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify entry is expired
	cachedResult, found = cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, cachedResult)

	// Give time for lazy cleanup
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, cache.Size())
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := New(5 * time.Minute)
	
	// Add multiple entries
	for i := 0; i < 5; i++ {
		key := CacheKey{
			URL:      "https://example.com",
			Strategy: "mobile",
			Locale:   string(rune('a' + i)),
		}
		result := &pagespeedonline.PagespeedApiPagespeedResponseV5{
			Id: "test-id",
		}
		cache.Set(key, result)
	}

	assert.Equal(t, 5, cache.Size())

	// Clear cache
	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

func TestInMemoryCache_DisabledCache(t *testing.T) {
	cache := New(0) // TTL of 0 disables cache
	
	key := CacheKey{
		URL:      "https://example.com",
		Strategy: "mobile",
	}
	
	result := &pagespeedonline.PagespeedApiPagespeedResponseV5{
		Id: "test-id",
	}

	// Verify set is ignored
	cache.Set(key, result)
	assert.Equal(t, 0, cache.Size())

	// Verify get returns nothing
	cachedResult, found := cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, cachedResult)
}

func TestInMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := New(5 * time.Minute)
	
	// Test concurrent reads and writes
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines * 2)

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := CacheKey{
				URL:      "https://example.com",
				Strategy: "mobile",
				Locale:   string(rune('a' + (id % 26))),
			}
			result := &pagespeedonline.PagespeedApiPagespeedResponseV5{
				Id: "test-id",
			}
			cache.Set(key, result)
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := CacheKey{
				URL:      "https://example.com",
				Strategy: "mobile",
				Locale:   string(rune('a' + (id % 26))),
			}
			cache.Get(key)
		}(i)
	}

	wg.Wait()
	
	// Verify cache is still functional
	assert.GreaterOrEqual(t, cache.Size(), 1)
	assert.LessOrEqual(t, cache.Size(), 26) // Max 26 unique keys
}

func TestInMemoryCache_NilResult(t *testing.T) {
	cache := New(5 * time.Minute)
	
	key := CacheKey{
		URL:      "https://example.com",
		Strategy: "mobile",
	}

	// Setting nil result should be ignored
	cache.Set(key, nil)
	assert.Equal(t, 0, cache.Size())

	cachedResult, found := cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, cachedResult)
}

func TestNewCacheKey(t *testing.T) {
	categories := []string{"performance", "seo"}
	key := NewCacheKey("https://example.com", "mobile", categories, "en", "campaign1", "source1")

	assert.Equal(t, "https://example.com", key.URL)
	assert.Equal(t, "mobile", key.Strategy)
	assert.Equal(t, []string{"performance", "seo"}, key.Categories)
	assert.Equal(t, "en", key.Locale)
	assert.Equal(t, "campaign1", key.Campaign)
	assert.Equal(t, "source1", key.Source)

	// Verify that modifying original categories doesn't affect cache key
	categories[0] = "modified"
	assert.Equal(t, []string{"performance", "seo"}, key.Categories)
}