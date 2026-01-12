package report

import (
	"fmt"
	"strings"

	"bedrock-performance/internal/config"
	"bedrock-performance/internal/types"
)

// ConsoleReporter handles real-time console output
type ConsoleReporter struct{}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

// PrintHeader prints the test header
func (c *ConsoleReporter) PrintHeader(cfg *config.Config) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("AWS Bedrock Performance Benchmark Tool")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Model: %s\n", cfg.Model.ID)
	fmt.Printf("Region: %s\n", cfg.AWS.Region)
	fmt.Printf("Prompt Size: %d characters\n", cfg.Test.PromptSize)
	fmt.Printf("Max Tokens: %d\n", cfg.Test.MaxTokens)
	fmt.Printf("Temperature: %.2f\n", cfg.Test.Temperature)
	fmt.Printf("Concurrency Range: %d -> %d (step: %d)\n",
		cfg.Concurrency.Start, cfg.Concurrency.End, cfg.Concurrency.Step)
	fmt.Printf("Duration per Level: %d seconds\n", cfg.Concurrency.DurationSeconds)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
}

// PrintSection prints a section header
func (c *ConsoleReporter) PrintSection(title string) {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf(">>> %s\n", title)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()
}

// PrintConcurrencyLevel prints the start of a new concurrency level test
func (c *ConsoleReporter) PrintConcurrencyLevel(level int) {
	fmt.Printf("\n[Concurrency Level: %d]\n", level)
	fmt.Println("Starting test...")
}

// PrintProgress prints progress during the test
func (c *ConsoleReporter) PrintProgress(stats *types.Stats, concurrency int) {
	fmt.Printf("  Progress: %d requests | Success: %d | Failures: %d | Req/s: %.2f | Tokens/s: %.2f\n",
		stats.TotalRequests,
		stats.SuccessCount,
		stats.FailureCount,
		stats.RequestsPerSecond,
		stats.TokenThroughput,
	)
}

// PrintStats prints detailed statistics for a completed test
func (c *ConsoleReporter) PrintStats(stats *types.Stats, concurrency int) {
	fmt.Println("\nResults:")
	fmt.Println(strings.Repeat("─", 80))

	// General stats
	fmt.Printf("  Total Requests:     %d\n", stats.TotalRequests)
	fmt.Printf("  Successful:         %d (%.2f%%)\n", stats.SuccessCount, stats.SuccessRate)
	fmt.Printf("  Failed:             %d\n", stats.FailureCount)
	fmt.Printf("  Duration:           %s\n", stats.Duration.Round(100))

	// Throughput
	fmt.Println("\n  Throughput:")
	fmt.Printf("    Requests/sec:     %.2f\n", stats.RequestsPerSecond)
	fmt.Printf("    Tokens/sec:       %.2f\n", stats.TokenThroughput)

	// Token stats
	fmt.Println("\n  Token Usage:")
	fmt.Printf("    Input Tokens:     %d\n", stats.TotalInputTokens)
	fmt.Printf("    Output Tokens:    %d\n", stats.TotalOutputTokens)
	fmt.Printf("    Total Tokens:     %d\n", stats.TotalTokens)

	// Latency stats
	if stats.SuccessCount > 0 {
		fmt.Println("\n  Latency (ms):")
		fmt.Printf("    Average:          %.2f\n", stats.AvgLatency)
		fmt.Printf("    Min:              %.2f\n", stats.MinLatency)
		fmt.Printf("    Max:              %.2f\n", stats.MaxLatency)
		fmt.Printf("    P50:              %.2f\n", stats.P50Latency)
		fmt.Printf("    P95:              %.2f\n", stats.P95Latency)
		fmt.Printf("    P99:              %.2f\n", stats.P99Latency)
	}

	// TTFT stats (if available)
	if stats.HasTTFT {
		fmt.Println("\n  Time to First Token (ms):")
		fmt.Printf("    Average:          %.2f\n", stats.AvgTTFT)
		fmt.Printf("    Min:              %.2f\n", stats.MinTTFT)
		fmt.Printf("    Max:              %.2f\n", stats.MaxTTFT)
		fmt.Printf("    P50:              %.2f\n", stats.P50TTFT)
		fmt.Printf("    P95:              %.2f\n", stats.P95TTFT)
		fmt.Printf("    P99:              %.2f\n", stats.P99TTFT)
	}

	// Error distribution
	if len(stats.ErrorsByType) > 0 {
		fmt.Println("\n  Error Distribution:")
		for errType, count := range stats.ErrorsByType {
			fmt.Printf("    %s: %d\n", errType, count)
		}
	}

	fmt.Println(strings.Repeat("─", 80))
}

// PrintReportSaved prints a message indicating the report was saved
func (c *ConsoleReporter) PrintReportSaved(filename string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Report saved to: %s\n", filename)
	fmt.Println(strings.Repeat("=", 80))
}

// PrintError prints an error message
func (c *ConsoleReporter) PrintError(err error) {
	fmt.Printf("\n[ERROR] %v\n", err)
}
