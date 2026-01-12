package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"bedrock-performance/internal/config"
	"bedrock-performance/internal/types"
)

// MarkdownReporter generates markdown reports
type MarkdownReporter struct {
	config *config.Config
}

// NewMarkdownReporter creates a new markdown reporter
func NewMarkdownReporter(cfg *config.Config) *MarkdownReporter {
	return &MarkdownReporter{
		config: cfg,
	}
}

// Generate generates the full markdown report
func (m *MarkdownReporter) Generate(allStats []*types.ConcurrencyLevelStats) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# AWS Bedrock Performance Benchmark Report\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Test Configuration
	m.writeConfiguration(&sb)

	// Overall Summary
	m.writeOverallSummary(&sb, allStats)

	// Detailed Results by Concurrency Level
	m.writeDetailedResults(&sb, allStats)

	// Latency Analysis
	m.writeLatencyAnalysis(&sb, allStats)

	// TTFT Analysis (if available)
	m.writeTTFTAnalysis(&sb, allStats)

	// Error Analysis
	m.writeErrorAnalysis(&sb, allStats)

	return sb.String()
}

// writeConfiguration writes the test configuration section
func (m *MarkdownReporter) writeConfiguration(sb *strings.Builder) {
	sb.WriteString("## Test Configuration\n\n")
	sb.WriteString("| Parameter | Value |\n")
	sb.WriteString("|-----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Model | %s |\n", m.config.Model.ID))
	sb.WriteString(fmt.Sprintf("| Region | %s |\n", m.config.AWS.Region))
	sb.WriteString(fmt.Sprintf("| Quota | %d |\n", m.config.Model.Quota))
	sb.WriteString(fmt.Sprintf("| Prompt Size | %d characters |\n", m.config.Test.PromptSize))
	sb.WriteString(fmt.Sprintf("| Max Tokens | %d |\n", m.config.Test.MaxTokens))
	sb.WriteString(fmt.Sprintf("| Temperature | %.2f |\n", m.config.Test.Temperature))
	sb.WriteString(fmt.Sprintf("| Streaming Enabled | %t |\n", m.config.Test.Streaming))
	sb.WriteString(fmt.Sprintf("| Non-Streaming Enabled | %t |\n", m.config.Test.NonStreaming))
	sb.WriteString(fmt.Sprintf("| Concurrency Range | %d - %d (step: %d) |\n",
		m.config.Concurrency.Start, m.config.Concurrency.End, m.config.Concurrency.Step))
	sb.WriteString(fmt.Sprintf("| Duration per Level | %d seconds |\n\n", m.config.Concurrency.DurationSeconds))
}

// writeOverallSummary writes the overall summary section
func (m *MarkdownReporter) writeOverallSummary(sb *strings.Builder, allStats []*types.ConcurrencyLevelStats) {
	sb.WriteString("## Overall Summary\n\n")

	totalRequests := 0
	totalSuccess := 0
	totalFailures := 0
	totalTokens := 0

	for _, stat := range allStats {
		totalRequests += stat.Stats.TotalRequests
		totalSuccess += stat.Stats.SuccessCount
		totalFailures += stat.Stats.FailureCount
		totalTokens += stat.Stats.TotalTokens
	}

	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(totalSuccess) / float64(totalRequests) * 100.0
	}

	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Requests | %d |\n", totalRequests))
	sb.WriteString(fmt.Sprintf("| Successful Requests | %d (%.2f%%) |\n", totalSuccess, successRate))
	sb.WriteString(fmt.Sprintf("| Failed Requests | %d |\n", totalFailures))
	sb.WriteString(fmt.Sprintf("| Total Tokens Processed | %d |\n\n", totalTokens))
}

// writeDetailedResults writes detailed results for each concurrency level
func (m *MarkdownReporter) writeDetailedResults(sb *strings.Builder, allStats []*types.ConcurrencyLevelStats) {
	sb.WriteString("## Detailed Results by Concurrency Level\n\n")

	sb.WriteString("| Concurrency | Requests | Success Rate | Req/s | Tokens/s | Avg Latency (ms) | P50 (ms) | P95 (ms) | P99 (ms) |\n")
	sb.WriteString("|-------------|----------|--------------|-------|----------|------------------|----------|----------|----------|\n")

	for _, stat := range allStats {
		s := stat.Stats
		sb.WriteString(fmt.Sprintf("| %d | %d | %.2f%% | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
			stat.ConcurrencyLevel,
			s.TotalRequests,
			s.SuccessRate,
			s.RequestsPerSecond,
			s.TokenThroughput,
			s.AvgLatency,
			s.P50Latency,
			s.P95Latency,
			s.P99Latency,
		))
	}
	sb.WriteString("\n")
}

// writeLatencyAnalysis writes latency analysis section
func (m *MarkdownReporter) writeLatencyAnalysis(sb *strings.Builder, allStats []*types.ConcurrencyLevelStats) {
	sb.WriteString("## Latency Analysis\n\n")
	sb.WriteString("### Latency Distribution by Concurrency Level\n\n")

	sb.WriteString("| Concurrency | Min (ms) | Avg (ms) | Max (ms) | P50 (ms) | P95 (ms) | P99 (ms) |\n")
	sb.WriteString("|-------------|----------|----------|----------|----------|----------|----------|\n")

	for _, stat := range allStats {
		s := stat.Stats
		if s.SuccessCount > 0 {
			sb.WriteString(fmt.Sprintf("| %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
				stat.ConcurrencyLevel,
				s.MinLatency,
				s.AvgLatency,
				s.MaxLatency,
				s.P50Latency,
				s.P95Latency,
				s.P99Latency,
			))
		}
	}
	sb.WriteString("\n")
}

// writeTTFTAnalysis writes TTFT analysis section (if available)
func (m *MarkdownReporter) writeTTFTAnalysis(sb *strings.Builder, allStats []*types.ConcurrencyLevelStats) {
	// Check if any stats have TTFT data
	hasTTFT := false
	for _, stat := range allStats {
		if stat.Stats.HasTTFT {
			hasTTFT = true
			break
		}
	}

	if !hasTTFT {
		return
	}

	sb.WriteString("## Time to First Token (TTFT) Analysis\n\n")
	sb.WriteString("### TTFT Distribution by Concurrency Level (Streaming Mode)\n\n")

	sb.WriteString("| Concurrency | Min (ms) | Avg (ms) | Max (ms) | P50 (ms) | P95 (ms) | P99 (ms) |\n")
	sb.WriteString("|-------------|----------|----------|----------|----------|----------|----------|\n")

	for _, stat := range allStats {
		s := stat.Stats
		if s.HasTTFT {
			sb.WriteString(fmt.Sprintf("| %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
				stat.ConcurrencyLevel,
				s.MinTTFT,
				s.AvgTTFT,
				s.MaxTTFT,
				s.P50TTFT,
				s.P95TTFT,
				s.P99TTFT,
			))
		}
	}
	sb.WriteString("\n")
}

// writeErrorAnalysis writes error analysis section
func (m *MarkdownReporter) writeErrorAnalysis(sb *strings.Builder, allStats []*types.ConcurrencyLevelStats) {
	// Aggregate errors across all concurrency levels
	allErrors := make(map[string]int)
	for _, stat := range allStats {
		for errType, count := range stat.Stats.ErrorsByType {
			allErrors[errType] += count
		}
	}

	if len(allErrors) == 0 {
		sb.WriteString("## Error Analysis\n\n")
		sb.WriteString("No errors occurred during the test. ✓\n\n")
		return
	}

	sb.WriteString("## Error Analysis\n\n")
	sb.WriteString("### Error Distribution\n\n")

	sb.WriteString("| Error Type | Count |\n")
	sb.WriteString("|------------|-------|\n")

	for errType, count := range allErrors {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", errType, count))
	}
	sb.WriteString("\n")

	// Error breakdown by concurrency level
	sb.WriteString("### Errors by Concurrency Level\n\n")
	sb.WriteString("| Concurrency | Total Errors | Error Types |\n")
	sb.WriteString("|-------------|--------------|-------------|\n")

	for _, stat := range allStats {
		if stat.Stats.FailureCount > 0 {
			errorTypes := []string{}
			for errType, count := range stat.Stats.ErrorsByType {
				errorTypes = append(errorTypes, fmt.Sprintf("%s(%d)", errType, count))
			}
			sb.WriteString(fmt.Sprintf("| %d | %d | %s |\n",
				stat.ConcurrencyLevel,
				stat.Stats.FailureCount,
				strings.Join(errorTypes, ", "),
			))
		}
	}
	sb.WriteString("\n")
}

// SaveToFile saves the report to a file
func (m *MarkdownReporter) SaveToFile(content string, filename string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
