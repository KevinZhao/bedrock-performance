package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bedrock-performance/internal/benchmark"
	"bedrock-performance/internal/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Create and run the benchmark
	runner := benchmark.NewRunner(cfg)

	allStats, err := runner.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}

	// Generate report
	if err := runner.GenerateReport(allStats); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate report: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nBenchmark completed successfully!")
}
