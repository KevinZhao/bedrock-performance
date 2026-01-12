package benchmark

import (
	"context"
	"fmt"
	"time"

	"bedrock-performance/internal/bedrock"
	"bedrock-performance/internal/config"
	"bedrock-performance/internal/report"
	"bedrock-performance/internal/types"
)

// Runner orchestrates the benchmark test
type Runner struct {
	config       *config.Config
	clientConfig *bedrock.ClientConfig
	console      *report.ConsoleReporter
}

// NewRunner creates a new benchmark runner
func NewRunner(cfg *config.Config) *Runner {
	clientConfig := &bedrock.ClientConfig{
		Region:      cfg.AWS.Region,
		AccessKey:   cfg.AWS.AccessKeyID,
		SecretKey:   cfg.AWS.SecretAccessKey,
		ModelID:     cfg.Model.ID,
		MaxTokens:   cfg.Test.MaxTokens,
		Temperature: cfg.Test.Temperature,
		ServiceTier: bedrock.ServiceTier(cfg.Test.ServiceTier),
	}

	console := report.NewConsoleReporter()

	return &Runner{
		config:       cfg,
		clientConfig: clientConfig,
		console:      console,
	}
}

// Run executes the benchmark test
func (r *Runner) Run(ctx context.Context) ([]*types.ConcurrencyLevelStats, error) {
	r.console.PrintHeader(r.config)

	var allStats []*types.ConcurrencyLevelStats
	prompt := GeneratePrompt(r.config.Test.PromptTemplate, r.config.Test.PromptSize)

	// Test streaming if enabled
	if r.config.Test.Streaming {
		r.console.PrintSection("Streaming Mode Test")
		stats, err := r.runConcurrencyTests(ctx, prompt, true)
		if err != nil {
			return nil, fmt.Errorf("streaming test failed: %w", err)
		}
		allStats = append(allStats, stats...)
	}

	// Test non-streaming if enabled
	if r.config.Test.NonStreaming {
		r.console.PrintSection("Non-Streaming Mode Test")
		stats, err := r.runConcurrencyTests(ctx, prompt, false)
		if err != nil {
			return nil, fmt.Errorf("non-streaming test failed: %w", err)
		}
		allStats = append(allStats, stats...)
	}

	return allStats, nil
}

// runConcurrencyTests runs tests with increasing concurrency levels
func (r *Runner) runConcurrencyTests(ctx context.Context, prompt string, streaming bool) ([]*types.ConcurrencyLevelStats, error) {
	var results []*types.ConcurrencyLevelStats

	for concurrency := r.config.Concurrency.Start; concurrency <= r.config.Concurrency.End; concurrency += r.config.Concurrency.Step {
		r.console.PrintConcurrencyLevel(concurrency)

		stats, err := r.runSingleConcurrencyLevel(ctx, prompt, streaming, concurrency)
		if err != nil {
			return nil, fmt.Errorf("concurrency level %d failed: %w", concurrency, err)
		}

		results = append(results, &types.ConcurrencyLevelStats{
			ConcurrencyLevel: concurrency,
			Stats:            stats,
		})

		r.console.PrintStats(stats, concurrency)
	}

	return results, nil
}

// runSingleConcurrencyLevel runs a test at a specific concurrency level
func (r *Runner) runSingleConcurrencyLevel(ctx context.Context, prompt string, streaming bool, concurrency int) (*types.Stats, error) {
	metrics := NewMetrics()

	// Create worker pool - each worker will create its own client
	pool := NewWorkerPool(r.clientConfig, metrics, streaming, prompt, concurrency)

	// Create a context with timeout
	testCtx, cancel := context.WithTimeout(ctx, time.Duration(r.config.Concurrency.DurationSeconds)*time.Second)
	defer cancel()

	// Start workers
	pool.Start(testCtx)

	// Create a ticker for progress updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Monitor progress
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-testCtx.Done():
				close(done)
				return
			case <-ticker.C:
				currentStats := metrics.GetCurrentStats()
				r.console.PrintProgress(currentStats, concurrency)
			}
		}
	}()

	// Wait for test to complete
	<-testCtx.Done()
	<-done

	// Stop all workers
	pool.Stop()

	// Finalize metrics
	metrics.Finalize()

	return metrics.ComputeStats(), nil
}

// GenerateReport generates the final benchmark report
func (r *Runner) GenerateReport(allStats []*types.ConcurrencyLevelStats) error {
	generator := report.NewMarkdownReporter(r.config)

	reportContent := generator.Generate(allStats)

	if err := generator.SaveToFile(reportContent, r.config.Output.ReportFile); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	r.console.PrintReportSaved(r.config.Output.ReportFile)

	return nil
}
