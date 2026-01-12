package benchmark

import (
	"math"
	"sort"
	"sync"
	"time"

	"bedrock-performance/internal/bedrock"
	"bedrock-performance/internal/types"
)

// Metrics collects and computes benchmark statistics
type Metrics struct {
	mu sync.Mutex

	// Raw data
	results       []*bedrock.InvokeResult
	startTime     time.Time
	endTime       time.Time

	// Counters
	totalRequests   int
	successCount    int
	failureCount    int
	totalInputTokens  int
	totalOutputTokens int

	// Error tracking
	errorsByType map[string]int

	// Latency data (in milliseconds)
	latencies []float64
	ttfts     []float64 // Only for streaming
}

// NewMetrics creates a new Metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		results:      make([]*bedrock.InvokeResult, 0),
		errorsByType: make(map[string]int),
		latencies:    make([]float64, 0),
		ttfts:        make([]float64, 0),
		startTime:    time.Now(),
	}
}

// AddResult adds a result to the metrics
func (m *Metrics) AddResult(result *bedrock.InvokeResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.results = append(m.results, result)
	m.totalRequests++

	if result.Success {
		m.successCount++
		m.totalInputTokens += result.InputTokens
		m.totalOutputTokens += result.OutputTokens

		// Record latency in milliseconds
		latencyMs := float64(result.Duration().Microseconds()) / 1000.0
		m.latencies = append(m.latencies, latencyMs)

		// Record TTFT if available (streaming only)
		if result.TTFT > 0 {
			ttftMs := float64(result.TTFT.Microseconds()) / 1000.0
			m.ttfts = append(m.ttfts, ttftMs)
		}
	} else {
		m.failureCount++
		if result.ErrorType != "" {
			m.errorsByType[result.ErrorType]++
		} else {
			m.errorsByType["UnknownError"]++
		}
	}
}

// Finalize marks the end of metric collection
func (m *Metrics) Finalize() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.endTime = time.Now()
}

// ComputeStats computes statistics from collected metrics
func (m *Metrics) ComputeStats() *types.Stats {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := &types.Stats{
		TotalRequests:     m.totalRequests,
		SuccessCount:      m.successCount,
		FailureCount:      m.failureCount,
		TotalInputTokens:  m.totalInputTokens,
		TotalOutputTokens: m.totalOutputTokens,
		TotalTokens:       m.totalInputTokens + m.totalOutputTokens,
		ErrorsByType:      make(map[string]int),
	}

	// Copy error map
	for k, v := range m.errorsByType {
		stats.ErrorsByType[k] = v
	}

	// Calculate success rate
	if m.totalRequests > 0 {
		stats.SuccessRate = float64(m.successCount) / float64(m.totalRequests) * 100.0
	}

	// Calculate duration
	if m.endTime.IsZero() {
		stats.Duration = time.Since(m.startTime)
	} else {
		stats.Duration = m.endTime.Sub(m.startTime)
	}

	durationSeconds := stats.Duration.Seconds()

	// Calculate throughput
	if durationSeconds > 0 {
		stats.RequestsPerSecond = float64(m.successCount) / durationSeconds
		stats.TokenThroughput = float64(stats.TotalTokens) / durationSeconds
	}

	// Calculate latency statistics
	if len(m.latencies) > 0 {
		sortedLatencies := make([]float64, len(m.latencies))
		copy(sortedLatencies, m.latencies)
		sort.Float64s(sortedLatencies)

		stats.AvgLatency = average(sortedLatencies)
		stats.MinLatency = sortedLatencies[0]
		stats.MaxLatency = sortedLatencies[len(sortedLatencies)-1]
		stats.P50Latency = percentile(sortedLatencies, 50)
		stats.P95Latency = percentile(sortedLatencies, 95)
		stats.P99Latency = percentile(sortedLatencies, 99)
	}

	// Calculate TTFT statistics if available
	if len(m.ttfts) > 0 {
		stats.HasTTFT = true
		sortedTTFTs := make([]float64, len(m.ttfts))
		copy(sortedTTFTs, m.ttfts)
		sort.Float64s(sortedTTFTs)

		stats.AvgTTFT = average(sortedTTFTs)
		stats.MinTTFT = sortedTTFTs[0]
		stats.MaxTTFT = sortedTTFTs[len(sortedTTFTs)-1]
		stats.P50TTFT = percentile(sortedTTFTs, 50)
		stats.P95TTFT = percentile(sortedTTFTs, 95)
		stats.P99TTFT = percentile(sortedTTFTs, 99)
	}

	return stats
}

// GetCurrentStats returns current statistics without finalizing
func (m *Metrics) GetCurrentStats() *types.Stats {
	return m.ComputeStats()
}

// Reset clears all collected metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.results = make([]*bedrock.InvokeResult, 0)
	m.totalRequests = 0
	m.successCount = 0
	m.failureCount = 0
	m.totalInputTokens = 0
	m.totalOutputTokens = 0
	m.errorsByType = make(map[string]int)
	m.latencies = make([]float64, 0)
	m.ttfts = make([]float64, 0)
	m.startTime = time.Now()
	m.endTime = time.Time{}
}

// average calculates the average of a slice of float64
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// percentile calculates the percentile of a sorted slice
func percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	if p <= 0 {
		return sortedValues[0]
	}
	if p >= 100 {
		return sortedValues[len(sortedValues)-1]
	}

	index := (p / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}
