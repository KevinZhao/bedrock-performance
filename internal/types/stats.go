package types

import "time"

// Stats contains computed statistics
type Stats struct {
	// General stats
	TotalRequests     int
	SuccessCount      int
	FailureCount      int
	SuccessRate       float64
	Duration          time.Duration

	// Token stats
	TotalInputTokens  int
	TotalOutputTokens int
	TotalTokens       int
	TokenThroughput   float64 // tokens per second

	// Latency stats (in milliseconds)
	AvgLatency float64
	MinLatency float64
	MaxLatency float64
	P50Latency float64
	P95Latency float64
	P99Latency float64

	// TTFT stats (in milliseconds, only for streaming)
	HasTTFT    bool
	AvgTTFT    float64
	MinTTFT    float64
	MaxTTFT    float64
	P50TTFT    float64
	P95TTFT    float64
	P99TTFT    float64

	// Throughput
	RequestsPerSecond float64

	// Errors
	ErrorsByType map[string]int
}

// ConcurrencyLevelStats tracks stats for a specific concurrency level
type ConcurrencyLevelStats struct {
	ConcurrencyLevel int
	Stats            *Stats
}
