# Lighthouse Results Caching Implementation Specification

## Overview

This specification describes the implementation of a simple in-memory caching mechanism for Google PageSpeed Insights (Lighthouse) results in the pagespeed_exporter. The cache will store results to reduce API calls and improve response times for repeated requests with identical configurations.

## Requirements

### Functional Requirements

1. **Cache Storage**: Store Lighthouse results in memory for configured duration
2. **Cache Key**: Uniquely identify cached results based on request configuration
3. **Cache Expiration**: Automatically expire cached entries after configured TTL
4. **Configuration**: Support cache duration configuration passed from main
5. **Cache Bypass**: Return cached results when available and valid, otherwise fetch fresh data

### Non-Functional Requirements

1. **Performance**: Minimal overhead for cache lookups
2. **Memory Management**: Efficient memory usage with lazy cleanup of expired entries
3. **Thread Safety**: Support concurrent access in parallel execution mode
4. **No External Dependencies**: Use only Go standard library

## Design

### Cache Key Structure

The cache key must uniquely identify a PageSpeed request configuration:

```go
type CacheKey struct {
    URL        string
    Strategy   string   // "mobile" or "desktop"
    Categories []string // sorted list of categories
    Locale     string
    Campaign   string
    Source     string
}
```

### Cache Entry Structure

```go
type CacheEntry struct {
    Key        CacheKey
    Result     *pagespeedpb.RunPagespeedResponse
    Timestamp  time.Time
    TTL        time.Duration
}
```

### Configuration

The cache TTL will be configured in main and passed to the cache instance:
- Environment variable: `PAGESPEED_CACHE_TTL`
- Format: Duration string (e.g., "15m", "1h", "30s")
- Default: "15m" (15 minutes)
- Disable cache: "0s" or "0"

### Implementation Components

#### 1. Cache Package Structure

The cache will be implemented in a separate `cache` package:

```
cache/
├── cache.go     # Interface and implementation
└── cache_test.go # Unit tests
```

#### 2. Cache Interface

```go
package cache

type ResultCache interface {
    Get(key CacheKey) (*pagespeedpb.RunPagespeedResponse, bool)
    Set(key CacheKey, result *pagespeedpb.RunPagespeedResponse)
    Clear()
}
```

#### 3. In-Memory Cache Implementation

```go
type InMemoryCache struct {
    mu         sync.RWMutex
    entries    map[string]*CacheEntry
    ttl        time.Duration
}

func New(ttl time.Duration) *InMemoryCache {
    return &InMemoryCache{
        entries: make(map[string]*CacheEntry),
        ttl:     ttl,
    }
}
```

#### 4. Integration Points

##### collector/scrape.go modifications:

1. Before making PageSpeed API call:
   - Generate cache key from request parameters
   - Check cache for existing valid entry
   - Return cached result if found

2. After successful PageSpeed API call:
   - Store result in cache with TTL

##### pagespeed_exporter.go modifications:

1. Parse `PAGESPEED_CACHE_TTL` environment variable
2. Initialize cache with configured TTL
3. Pass cache instance to collector factory

##### handler/probe.go modifications:

1. Use same cache instance for probe endpoint
2. Ensure cache key generation is consistent

### Cache Eviction Strategy

1. **Time-based expiration**: Primary mechanism based on TTL
2. **Lazy deletion**: Check expiration on access

## Implementation Steps

1. **Create cache package** (`cache/cache.go`)
   - Define interfaces and structures
   - Implement in-memory cache with TTL support

2. **Add configuration parsing** (`pagespeed_exporter.go`)
   - Parse PAGESPEED_CACHE_TTL environment variable
   - Set default 15-minute TTL
   - Pass cache instance to collector factory and probe handler

3. **Integrate with scraper** (`collector/scrape.go`)
   - Add cache checks before API calls
   - Store results after successful API calls

4. **Update tests**
   - Add unit tests for cache implementation
   - Add integration tests for cached responses

5. **Documentation updates**
   - Update README.md with cache configuration
   - Add cache behavior to CLAUDE.md

## Testing Strategy

1. **Unit Tests**:
   - Cache key generation
   - TTL expiration
   - Concurrent access
   - Eviction logic

2. **Integration Tests**:
   - End-to-end caching behavior
   - Configuration parsing

## Security Considerations

1. **Memory Limits**: Implement maximum cache size to prevent memory exhaustion
2. **Cache Poisoning**: Validate cached data integrity
3. **Key Collision**: Ensure cache key uniqueness

## Future Enhancements

1. **Distributed Cache**: Support for Redis/Memcached
2. **Cache Warming**: Pre-populate cache on startup
3. **Selective Invalidation**: Manual cache invalidation endpoints
4. **Compression**: Compress cached responses to reduce memory usage