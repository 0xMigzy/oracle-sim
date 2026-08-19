package queue

import (
    "context"
    "sync"
    "time"
)

// Job represents a single oracle request.
type Job struct {
    ID        string
    Pair      string  // e.g., "ETH/USD"
    CreatedAt time.Time
    Result    chan Result
}

// Result represents the outcome of processing a Job.
type Result struct {
    Price float64
    Error error
}

// Queue is a thread-safe in-memory job queue.
type Queue struct {
    mu       sync.Mutex
    jobs     chan *Job
    capacity int
}

// New creates a new Queue with the given capacity.
func New(capacity int) *Queue {
    return &Queue{
        jobs:     make(chan *Job, capacity),
        capacity: capacity,
    }
}

// Submit adds a job to the queue. Returns an error if the queue is full.
func (q *Queue) Submit(ctx context.Context, job *Job) error {
    select {
    case q.jobs <- job:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// Next returns the next job from the queue. Blocks until a job is available
// or the context is cancelled.
func (q *Queue) Next(ctx context.Context) (*Job, error) {
    select {
    case job := <-q.jobs:
        return job, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// Depth returns the current number of jobs in the queue.
func (q *Queue) Depth() int {
    return len(q.jobs)
}