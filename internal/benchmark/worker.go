package benchmark

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"bedrock-performance/internal/bedrock"
)

// WorkerPool manages a pool of workers for concurrent testing
type WorkerPool struct {
	clientConfig *bedrock.ClientConfig
	metrics      *Metrics
	streaming    bool
	prompt       string
	workerCount  int
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
// Each worker will create its own client to avoid connection pool contention
func NewWorkerPool(clientConfig *bedrock.ClientConfig, metrics *Metrics, streaming bool, prompt string, workerCount int) *WorkerPool {
	return &WorkerPool{
		clientConfig: clientConfig,
		metrics:      metrics,
		streaming:    streaming,
		prompt:       prompt,
		workerCount:  workerCount,
		stopChan:     make(chan struct{}),
	}
}

// Start starts all workers in the pool
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

// Stop signals all workers to stop and waits for them to finish
func (wp *WorkerPool) Stop() {
	close(wp.stopChan)
	wp.wg.Wait()
}

// worker is the main worker loop
// Each worker creates its own client to avoid connection pool bottleneck
func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()

	// Create a dedicated client for this worker to avoid connection pool contention
	client := bedrock.NewClientFromConfig(wp.clientConfig)

	for {
		// Check if we should stop before starting a new request
		select {
		case <-wp.stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		// Execute one request with independent context
		// Use context.Background() so the request won't be canceled by test timeout
		// This allows in-flight requests to complete naturally even after test window expires
		var result *bedrock.InvokeResult
		if wp.streaming {
			result = client.InvokeStreaming(context.Background(), wp.prompt)
		} else {
			result = client.InvokeNonStreaming(context.Background(), wp.prompt)
		}

		// Record the result
		wp.metrics.AddResult(result)
	}
}

// GeneratePrompt generates a prompt of approximately the specified size
func GeneratePrompt(template string, size int) string {
	if template == "" {
		template = "Please write a detailed explanation about artificial intelligence, " +
			"covering its history, applications, and future prospects. " +
			"Make your response approximately {size} characters long."
	}

	// Replace {size} placeholder if present
	prompt := strings.ReplaceAll(template, "{size}", fmt.Sprintf("%d", size))

	// If the prompt is already large enough, return it
	if len(prompt) >= size {
		return prompt[:size]
	}

	// Pad the prompt to reach the desired size
	padding := strings.Repeat("Please provide more detailed information. ", (size-len(prompt))/45+1)
	prompt = prompt + " " + padding

	// Trim to exact size
	if len(prompt) > size {
		prompt = prompt[:size]
	}

	return prompt
}
