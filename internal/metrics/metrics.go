package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// RequestsTotal counts the total number of oracle requests.
var RequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "oracle_requests_total",
        Help: "Total number of oracle requests by pair and status.",
    },
    []string{"pair", "status"}, // labels: e.g., status="success" or status="error"
)

// RequestDuration measures the duration of oracle requests.
var RequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "oracle_request_duration_seconds",
        Help:    "Histogram of oracle request durations by pair.",
        Buckets: prometheus.DefBuckets, // default buckets work for sub-second to multi-second
    },
    []string{"pair"},
)

// QueueDepth tracks the current queue depth.
var QueueDepth = promauto.NewGauge(
    prometheus.GaugeOpts{
        Name: "oracle_queue_depth",
        Help: "Current depth of the oracle job queue.",
    },
)