package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/dunamismax/MTG-Card-Bot/pkg/errors"
	"github.com/dunamismax/MTG-Card-Bot/pkg/logging"
	"github.com/dunamismax/MTG-Card-Bot/pkg/scryfall"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Card        *scryfall.Card
	Timestamp   time.Time
	AccessCount int64
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired(ttl time.Duration) bool {
	return time.Since(e.Timestamp) > ttl
}

// CardCache provides caching functionality for MTG cards
type CardCache struct {
	entries   map[string]*CacheEntry
	mutex     sync.RWMutex
	ttl       time.Duration
	maxSize   int
	hits      int64
	misses    int64
	evictions int64
}

// NewCardCache creates a new card cache with specified TTL and max size
func NewCardCache(ttl time.Duration, maxSize int) *CardCache {
	cache := &CardCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// normalizeKey normalizes card names for consistent cache keys
func normalizeKey(cardName string) string {
	// Convert to lowercase and remove extra spaces
	normalized := strings.ToLower(strings.TrimSpace(cardName))
	// Replace multiple spaces with single space
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

// Get retrieves a card from the cache
func (c *CardCache) Get(cardName string) (*scryfall.Card, bool) {
	start := time.Now()
	key := normalizeKey(cardName)

	c.mutex.RLock()
	entry, exists := c.entries[key]
	c.mutex.RUnlock()

	if !exists {
		c.mutex.Lock()
		c.misses++
		c.mutex.Unlock()

		logging.LogCacheOperation("get", key, false, time.Since(start).Nanoseconds())
		return nil, false
	}

	if entry.IsExpired(c.ttl) {
		// Remove expired entry
		c.mutex.Lock()
		delete(c.entries, key)
		c.misses++
		c.mutex.Unlock()

		logging.LogCacheOperation("get", key, false, time.Since(start).Nanoseconds())
		return nil, false
	}

	// Update access count
	c.mutex.Lock()
	entry.AccessCount++
	c.hits++
	c.mutex.Unlock()

	logging.LogCacheOperation("get", key, true, time.Since(start).Nanoseconds())
	return entry.Card, true
}

// Set stores a card in the cache
func (c *CardCache) Set(cardName string, card *scryfall.Card) error {
	if card == nil {
		return errors.NewValidationError("cannot cache nil card")
	}

	start := time.Now()
	key := normalizeKey(cardName)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict entries to make space
	if len(c.entries) >= c.maxSize {
		if err := c.evictLRU(); err != nil {
			logging.LogCacheOperation("set", key, false, time.Since(start).Nanoseconds())
			return errors.NewCacheError("failed to evict entry for new cache item", err)
		}
	}

	entry := &CacheEntry{
		Card:        card,
		Timestamp:   time.Now(),
		AccessCount: 1,
	}

	c.entries[key] = entry

	logging.LogCacheOperation("set", key, true, time.Since(start).Nanoseconds())
	return nil
}

// evictLRU removes the least recently used entry (must be called with mutex locked)
func (c *CardCache) evictLRU() error {
	if len(c.entries) == 0 {
		return nil
	}

	var oldestKey string
	var oldestTime time.Time
	var minAccessCount int64 = -1

	// Find entry with oldest timestamp and lowest access count
	for key, entry := range c.entries {
		if minAccessCount == -1 || entry.AccessCount < minAccessCount ||
			(entry.AccessCount == minAccessCount && entry.Timestamp.Before(oldestTime)) {
			oldestKey = key
			oldestTime = entry.Timestamp
			minAccessCount = entry.AccessCount
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.evictions++
	}

	return nil
}

// Size returns the current number of entries in the cache
func (c *CardCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.entries)
}

// Stats returns cache statistics
func (c *CardCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.hits + c.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	return CacheStats{
		Size:      len(c.entries),
		MaxSize:   c.maxSize,
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		HitRate:   hitRate,
		TTL:       c.ttl,
	}
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	Size      int           `json:"size"`
	MaxSize   int           `json:"max_size"`
	Hits      int64         `json:"hits"`
	Misses    int64         `json:"misses"`
	Evictions int64         `json:"evictions"`
	HitRate   float64       `json:"hit_rate_percent"`
	TTL       time.Duration `json:"ttl"`
}

// Clear removes all entries from the cache
func (c *CardCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.hits = 0
	c.misses = 0
	c.evictions = 0

	logging.Debug("Cache cleared")
}

// cleanupLoop periodically removes expired entries
func (c *CardCache) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2) // Run cleanup at half the TTL interval
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries from the cache
func (c *CardCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiredKeys := make([]string, 0)

	for key, entry := range c.entries {
		if entry.IsExpired(c.ttl) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.entries, key)
	}

	if len(expiredKeys) > 0 {
		logging.Debug("Cache cleanup completed", "expired_entries", len(expiredKeys))
	}
}

// GetOrSet retrieves a card from cache or executes the provided function to get and cache it
func (c *CardCache) GetOrSet(cardName string, getter func(string) (*scryfall.Card, error)) (*scryfall.Card, error) {
	// First try to get from cache
	if card, found := c.Get(cardName); found {
		return card, nil
	}

	// Not in cache, execute getter function
	card, err := getter(cardName)
	if err != nil {
		return nil, err
	}

	// Store in cache (ignore cache errors, they shouldn't fail the main operation)
	if cacheErr := c.Set(cardName, card); cacheErr != nil {
		logging.Error("Failed to cache card", "card_name", cardName, "error", cacheErr)
	}

	return card, nil
}
