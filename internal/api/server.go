package api

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/0xmigzy/oracle-sim/internal/queue"
    "github.com/0xmigzy/oracle-sim/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

)

// Server wraps an HTTP server that exposes the oracle service.
type Server struct {
    queue *queue.Queue
    srv   *http.Server
}

// NewServer creates a new Server bound to the given address.
func NewServer(q *queue.Queue, addr string) *Server {
    s := &Server{queue: q}

    mux := http.NewServeMux()
    mux.HandleFunc("/oracle/request", s.handleOracleRequest)
    mux.HandleFunc("/oracle/health", s.handleHealth)
    mux.HandleFunc("/oracle/queue/depth", s.handleQueueDepth)
	mux.Handle("/metrics", promhttp.Handler())

    s.srv = &http.Server{
        Addr:         addr,
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    return s
}

// Start launches the HTTP server. Returns an error if the server fails to start.
func (s *Server) Start() error {
    return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
    return s.srv.Shutdown(ctx)
}

// OracleRequestPayload is the JSON body for an oracle request.
type OracleRequestPayload struct {
    Pair string `json:"pair"`
}

// OracleResponse is the JSON body for an oracle response.
type OracleResponse struct {
    Pair  string  `json:"pair"`
    Price float64 `json:"price,omitempty"`
    Error string  `json:"error,omitempty"`
}

// handleOracleRequest accepts an oracle request and returns the result.
func (s *Server) handleOracleRequest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var payload OracleRequestPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    if payload.Pair == "" {
        http.Error(w, "pair is required", http.StatusBadRequest)
        return
    }

	  start := time.Now()

    // Create the job
    job := &queue.Job{
        ID:        fmt.Sprintf("job-%d", time.Now().UnixNano()),
        Pair:      payload.Pair,
        CreatedAt: time.Now(),
        Result:    make(chan queue.Result, 1),
    }

    // Submit to queue with a timeout
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    if err := s.queue.Submit(ctx, job); err != nil {
        http.Error(w, "queue submit failed: "+err.Error(), http.StatusServiceUnavailable)
        return
    }

    // Wait for the result
    var result queue.Result
    select {
    case result = <-job.Result:
    case <-ctx.Done():
        http.Error(w, "request timed out", http.StatusGatewayTimeout)
        return
    }

	duration := time.Since(start).Seconds()
    metrics.RequestDuration.WithLabelValues(payload.Pair).Observe(duration)

    // Build response
    response := OracleResponse{Pair: payload.Pair}
    if result.Error != nil {
        response.Error = result.Error.Error()
        metrics.RequestsTotal.WithLabelValues(payload.Pair, "error").Inc()
        w.WriteHeader(http.StatusInternalServerError)
    } else {
        response.Price = result.Price
        metrics.RequestsTotal.WithLabelValues(payload.Pair, "success").Inc()
        w.WriteHeader(http.StatusOK)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleQueueDepth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]int{"depth": s.queue.Depth()})
}