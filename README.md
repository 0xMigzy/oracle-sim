# oracle-sim

A simulated price oracle service written in Go. It exposes an HTTP API for requesting asset prices, processes requests through an in-memory job queue and worker pool, and exports Prometheus metrics for observability.

This project is meant as a learning/reference implementation of common backend patterns: request queuing, worker pools, graceful shutdown, and metrics instrumentation — not a real price feed.

## Features

- **HTTP API** for submitting oracle price requests
- **In-memory job queue** with bounded capacity
- **Worker pool** that simulates fetching/aggregating prices with realistic latency and occasional failures
- **Prometheus metrics** for request counts, durations, and queue depth
- **Graceful shutdown** on `SIGINT` / `SIGTERM`

## Project structure

```
oracle-sim/
├── go.mod
├── go.sum
├── main.go
├── internal/
│   ├── api/
│   │   └── server.go       # HTTP handlers
│   ├── queue/
│   │   └── queue.go        # Thread-safe job queue
│   ├── worker/
│   │   └── worker.go       # Worker pool that processes jobs
│   └── metrics/
│       └── metrics.go      # Prometheus metrics definitions
└── README.md
```

## Requirements

- Go 1.21+ (check with `go version`)

## Getting started

Clone the repo and run the service:

```bash
git clone https://github.com/YOUR_USERNAME/oracle-sim.git
cd oracle-sim
go run main.go
```

The server starts on `:8080` by default.

## API

### `POST /oracle/request`

Request a simulated price for a trading pair.

**Request body:**

```json
{ "pair": "ETH/USD" }
```

**Example:**

```bash
curl -X POST http://localhost:8080/oracle/request \
  -H 'Content-Type: application/json' \
  -d '{"pair":"ETH/USD"}'
```

**Response:**

```json
{ "pair": "ETH/USD", "price": 3512.44 }
```

Supported pairs (with fallback for unknown pairs): `ETH/USD`, `BTC/USD`, `USDC/USD`, `WBTC/USD`.

### `GET /oracle/health`

Simple health check.

```bash
curl http://localhost:8080/oracle/health
```

### `GET /oracle/queue/depth`

Returns the current number of jobs waiting in the queue.

```bash
curl http://localhost:8080/oracle/queue/depth
```

### `GET /metrics`

Prometheus metrics endpoint.

```bash
curl http://localhost:8080/metrics | grep oracle
```

Exposed metrics:

| Metric | Type | Description |
|---|---|---|
| `oracle_requests_total` | Counter | Total requests, labeled by `pair` and `status` |
| `oracle_request_duration_seconds` | Histogram | Request duration, labeled by `pair` |
| `oracle_queue_depth` | Gauge | Current queue depth |

## Configuration

Currently configured via constants in `main.go`:

- Queue capacity: `100`
- Worker pool size: `5`
- HTTP address: `:8080`

## How it works

1. A client submits a price request via `POST /oracle/request`.
2. The request is wrapped in a `Job` and pushed onto the queue.
3. One of the worker goroutines picks up the job, simulates a fetch/aggregation delay (50–500ms), and occasionally simulates a failure (~5% of the time).
4. The result is sent back through a per-job result channel and returned to the HTTP caller.
5. Metrics are recorded for every request (duration, success/error count), and queue depth is sampled once per second.

## License

MIT (or update this to whatever you prefer).
