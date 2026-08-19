package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/0xmigzy/oracle-sim/internal/api"
    "github.com/0xmigzy/oracle-sim/internal/metrics"
    "github.com/0xmigzy/oracle-sim/internal/queue"
    "github.com/0xmigzy/oracle-sim/internal/worker"
)

func main() {
    log.Println("starting oracle-sim")

    // Setup
    q := queue.New(100) // buffer 100 jobs
    pool := worker.NewPool(q, 5) // 5 worker goroutines
    server := api.NewServer(q, ":8080")

    // Root context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start the worker pool
    go pool.Start(ctx)

    // Start the gauge updater
    go updateQueueDepthGauge(ctx, q)

    // Start the HTTP server
    go func() {
        log.Println("HTTP server listening on :8080")
        if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
            log.Fatalf("server error: %v", err)
        }
    }()

    // Wait for shutdown signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    log.Println("shutdown signal received")

    // Graceful shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
    }

    cancel() // signals workers and gauge updater to stop
    log.Println("oracle-sim stopped")
}

func updateQueueDepthGauge(ctx context.Context, q *queue.Queue) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            metrics.QueueDepth.Set(float64(q.Depth()))
        case <-ctx.Done():
            return
        }
    }
}