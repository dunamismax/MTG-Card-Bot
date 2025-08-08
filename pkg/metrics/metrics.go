package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/dunamismax/MTG-Card-Bot/pkg/errors"
)

// Metrics holds all application metrics
type Metrics struct {
	// Command metrics
	CommandsTotal      int64
	CommandsSuccessful int64
	CommandsFailed     int64
	CommandsPerSecond  float64

	// API metrics
	APIRequestsTotal      int64
	APIRequestsSuccessful int64
	APIRequestsFailed     int64
	APIRequestsPerSecond  float64
	APIResponseTimeSum    int64 // in milliseconds
	APIResponseCount      int64

	// Error metrics by type
	ErrorsByType map[errors.ErrorType]int64

	// Cache metrics
	CacheHits   int64
	CacheMisses int64
	CacheSize   int64

	// Bot metrics
	BotStartTime time.Time

	// Rate tracking
	commandWindow *RateWindow
	apiWindow     *RateWindow
	mutex         sync.RWMutex
}

// RateWindow tracks events within a time window for rate calculations
type RateWindow struct {
	events []time.Time
	window time.Duration
	mutex  sync.Mutex
}

// NewRateWindow creates a new rate tracking window
func NewRateWindow(window time.Duration) *RateWindow {
	return &RateWindow{
		events: make([]time.Time, 0),
		window: window,
	}
}

// Add records an event timestamp
func (rw *RateWindow) Add(timestamp time.Time) {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()

	// Add new event
	rw.events = append(rw.events, timestamp)

	// Remove events outside the window
	cutoff := timestamp.Add(-rw.window)
	validEvents := make([]time.Time, 0, len(rw.events))

	for _, event := range rw.events {
		if event.After(cutoff) {
			validEvents = append(validEvents, event)
		}
	}

	rw.events = validEvents
}

// Rate calculates the current rate per second
func (rw *RateWindow) Rate() float64 {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()

	if len(rw.events) == 0 {
		return 0.0
	}

	// Remove expired events
	now := time.Now()
	cutoff := now.Add(-rw.window)
	validEvents := 0

	for _, event := range rw.events {
		if event.After(cutoff) {
			validEvents++
		}
	}

	// Calculate rate per second
	windowSeconds := rw.window.Seconds()
	return float64(validEvents) / windowSeconds
}

var globalMetrics *Metrics
var once sync.Once

// Initialize sets up the global metrics instance
func Initialize() *Metrics {
	once.Do(func() {
		globalMetrics = &Metrics{
			ErrorsByType:  make(map[errors.ErrorType]int64),
			BotStartTime:  time.Now(),
			commandWindow: NewRateWindow(60 * time.Second), // 1-minute window
			apiWindow:     NewRateWindow(60 * time.Second), // 1-minute window
		}
	})
	return globalMetrics
}

// Get returns the global metrics instance
func Get() *Metrics {
	if globalMetrics == nil {
		return Initialize()
	}
	return globalMetrics
}

// IncrementCommands increments command counters
func (m *Metrics) IncrementCommands(successful bool) {
	now := time.Now()
	atomic.AddInt64(&m.CommandsTotal, 1)

	if successful {
		atomic.AddInt64(&m.CommandsSuccessful, 1)
	} else {
		atomic.AddInt64(&m.CommandsFailed, 1)
	}

	m.commandWindow.Add(now)
	m.CommandsPerSecond = m.commandWindow.Rate()
}

// IncrementAPIRequests increments API request counters
func (m *Metrics) IncrementAPIRequests(successful bool, responseTimeMs int64) {
	now := time.Now()
	atomic.AddInt64(&m.APIRequestsTotal, 1)

	if successful {
		atomic.AddInt64(&m.APIRequestsSuccessful, 1)
	} else {
		atomic.AddInt64(&m.APIRequestsFailed, 1)
	}

	// Track response time
	atomic.AddInt64(&m.APIResponseTimeSum, responseTimeMs)
	atomic.AddInt64(&m.APIResponseCount, 1)

	m.apiWindow.Add(now)
	m.APIRequestsPerSecond = m.apiWindow.Rate()
}

// IncrementError increments error counter by type
func (m *Metrics) IncrementError(errorType errors.ErrorType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ErrorsByType[errorType]++
}

// UpdateCacheStats updates cache-related metrics
func (m *Metrics) UpdateCacheStats(hits, misses, size int64) {
	atomic.StoreInt64(&m.CacheHits, hits)
	atomic.StoreInt64(&m.CacheMisses, misses)
	atomic.StoreInt64(&m.CacheSize, size)
}

// GetAverageResponseTime calculates the average API response time
func (m *Metrics) GetAverageResponseTime() float64 {
	responseTimeSum := atomic.LoadInt64(&m.APIResponseTimeSum)
	responseCount := atomic.LoadInt64(&m.APIResponseCount)

	if responseCount == 0 {
		return 0.0
	}

	return float64(responseTimeSum) / float64(responseCount)
}

// GetUptime returns the bot uptime
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.BotStartTime)
}

// GetSuccessRate calculates the command success rate as a percentage
func (m *Metrics) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&m.CommandsTotal)
	if total == 0 {
		return 0.0
	}

	successful := atomic.LoadInt64(&m.CommandsSuccessful)
	return (float64(successful) / float64(total)) * 100.0
}

// GetAPISuccessRate calculates the API success rate as a percentage
func (m *Metrics) GetAPISuccessRate() float64 {
	total := atomic.LoadInt64(&m.APIRequestsTotal)
	if total == 0 {
		return 0.0
	}

	successful := atomic.LoadInt64(&m.APIRequestsSuccessful)
	return (float64(successful) / float64(total)) * 100.0
}

// GetCacheHitRate calculates the cache hit rate as a percentage
func (m *Metrics) GetCacheHitRate() float64 {
	hits := atomic.LoadInt64(&m.CacheHits)
	misses := atomic.LoadInt64(&m.CacheMisses)
	total := hits + misses

	if total == 0 {
		return 0.0
	}

	return (float64(hits) / float64(total)) * 100.0
}

// Summary returns a comprehensive metrics summary
type Summary struct {
	// Command statistics
	CommandsTotal      int64   `json:"commands_total"`
	CommandsSuccessful int64   `json:"commands_successful"`
	CommandsFailed     int64   `json:"commands_failed"`
	CommandsPerSecond  float64 `json:"commands_per_second"`
	CommandSuccessRate float64 `json:"command_success_rate_percent"`

	// API statistics
	APIRequestsTotal      int64   `json:"api_requests_total"`
	APIRequestsSuccessful int64   `json:"api_requests_successful"`
	APIRequestsFailed     int64   `json:"api_requests_failed"`
	APIRequestsPerSecond  float64 `json:"api_requests_per_second"`
	APISuccessRate        float64 `json:"api_success_rate_percent"`
	AverageResponseTime   float64 `json:"average_response_time_ms"`

	// Cache statistics
	CacheHits    int64   `json:"cache_hits"`
	CacheMisses  int64   `json:"cache_misses"`
	CacheSize    int64   `json:"cache_size"`
	CacheHitRate float64 `json:"cache_hit_rate_percent"`

	// Error statistics
	ErrorsByType map[errors.ErrorType]int64 `json:"errors_by_type"`

	// System statistics
	UptimeSeconds float64 `json:"uptime_seconds"`
	BotStartTime  string  `json:"bot_start_time"`
}

// GetSummary returns a comprehensive metrics summary
func (m *Metrics) GetSummary() Summary {
	m.mutex.RLock()
	errorsByType := make(map[errors.ErrorType]int64)
	for k, v := range m.ErrorsByType {
		errorsByType[k] = v
	}
	m.mutex.RUnlock()

	return Summary{
		CommandsTotal:         atomic.LoadInt64(&m.CommandsTotal),
		CommandsSuccessful:    atomic.LoadInt64(&m.CommandsSuccessful),
		CommandsFailed:        atomic.LoadInt64(&m.CommandsFailed),
		CommandsPerSecond:     m.CommandsPerSecond,
		CommandSuccessRate:    m.GetSuccessRate(),
		APIRequestsTotal:      atomic.LoadInt64(&m.APIRequestsTotal),
		APIRequestsSuccessful: atomic.LoadInt64(&m.APIRequestsSuccessful),
		APIRequestsFailed:     atomic.LoadInt64(&m.APIRequestsFailed),
		APIRequestsPerSecond:  m.APIRequestsPerSecond,
		APISuccessRate:        m.GetAPISuccessRate(),
		AverageResponseTime:   m.GetAverageResponseTime(),
		CacheHits:             atomic.LoadInt64(&m.CacheHits),
		CacheMisses:           atomic.LoadInt64(&m.CacheMisses),
		CacheSize:             atomic.LoadInt64(&m.CacheSize),
		CacheHitRate:          m.GetCacheHitRate(),
		ErrorsByType:          errorsByType,
		UptimeSeconds:         m.GetUptime().Seconds(),
		BotStartTime:          m.BotStartTime.Format(time.RFC3339),
	}
}

// RecordCommand is a convenience function to record command execution
func RecordCommand(successful bool) {
	Get().IncrementCommands(successful)
}

// RecordAPIRequest is a convenience function to record API requests
func RecordAPIRequest(successful bool, responseTimeMs int64) {
	Get().IncrementAPIRequests(successful, responseTimeMs)
}

// RecordError is a convenience function to record errors
func RecordError(err error) {
	if mtgErr, ok := err.(*errors.MTGError); ok {
		Get().IncrementError(mtgErr.Type)
	} else {
		Get().IncrementError(errors.ErrorTypeInternal)
	}
}
