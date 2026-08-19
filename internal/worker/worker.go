package worker

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"

    "github.com/0xmigzy/oracle-sim/internal/queue"
)

// Pool is a worker pool that processes jobs from a Queue.
type Pool struct {
    queue *queue.Queue
    size  int
    wg    sync.WaitGroup
}

// NewPool creates a new worker pool of the given size.
func NewPool(q *queue.Queue, size int) *Pool {
    return &Pool{
        queue: q,
        size:  size,
    }
}

// Start launches the worker pool. Blocks until the context is cancelled.
func (p *Pool) Start(ctx context.Context) {
    for i := 0; i < p.size; i++ {
        p.wg.Add(1)
        go p.workerLoop(ctx, i)
    }
    p.wg.Wait()
}

// workerLoop runs a single worker, fetching and processing jobs until the context is cancelled.
func (p *Pool) workerLoop(ctx context.Context, workerID int) {
    defer p.wg.Done()

    for {
        job, err := p.queue.Next(ctx)
        if err != nil {
            // Context cancelled — shut down gracefully
            return
        }

        result := p.processJob(ctx, job)

        // Send result back through the job's result channel
        select {
        case job.Result <- result:
        case <-ctx.Done():
            return
        }
    }
}

// processJob simulates fetching a price from external sources.
// In a real oracle, this would call exchange APIs, run consensus, etc.
func (p *Pool) processJob(ctx context.Context, job *queue.Job) queue.Result {
    // Simulate latency — real oracles take 50-500ms to fetch and aggregate
    select {
    case <-time.After(time.Duration(50+rand.Intn(450)) * time.Millisecond):
    case <-ctx.Done():
        return queue.Result{Error: ctx.Err()}
    }

    // Simulate occasional failures (5% rate)
    if rand.Intn(100) < 5 {
        return queue.Result{Error: fmt.Errorf("simulated source failure")}
    }

    // Simulate a price based on the pair
    price := simulatePrice(job.Pair)
    return queue.Result{Price: price}
}

// simulatePrice returns a fake price for the given pair.
func simulatePrice(pair string) float64 {
    base := map[string]float64{
        "ETH/USD":  3500.0,
        "BTC/USD":  65000.0,
        "USDC/USD": 1.0,
        "WBTC/USD": 64900.0,
    }
    p, ok := base[pair]
    if !ok {
        return 100.0 // default for unknown pairs
    }
    // Add ±0.5% noise
    noise := (rand.Float64() - 0.5) * 0.01
    return p * (1 + noise)
}