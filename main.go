// Package main provides a service to shorten URLs with features including HTTP handlers, BoltDB storage, JSON APIs, analytics, user authentication, and rate limiting. The service generates short aliases for long URLs, tracks access metrics, and supports secure user management.
// The service uses Gorilla Mux for routing HTTP requests. Routes include creating short URLs, redirecting to original URLs, retrieving analytics, and user authentication endpoints. Middleware chains handle logging, authentication, and rate limiting.
// Storage is implemented using BoltDB for persistence. Two separate buckets are used: one for URL mappings (short-to-original) and one for analytics data (timestamps, IPs, user agents). In-memory maps are used as caches to reduce BoltDB access latency.
// URL data models include the original URL, short alias, creation timestamp, expiration time (optional), and owner ID (for authenticated users). Analytics models store the short alias, access timestamp, client IP, user agent, and referrer.
// The CreateShortURL HTTP handler accepts JSON payloads containing the original URL and optional expiration. It generates a unique short alias (base62 encoded, 6-8 characters), stores it in BoltDB, and returns the shortened URL in JSON format. Validation ensures URL correctness and optional JWT-based user ownership.
// The Redirect HTTP handler looks up the short alias in BoltDB or cache. If found, it returns a 302 redirect to the original URL. Concurrent access is managed using read/write mutexes for in-memory maps. Missing aliases return a 404 JSON error.
// The Analytics HTTP handler retrieves access data for a short URL, filtered by optional time ranges or user. Requires authentication via JWT in the Authorization header. Data is paginated and returned as JSON with click counts, unique visitors, and time-series metrics.
// User authentication is implemented using JWT tokens. The /login endpoint validates credentials against a BoltDB user bucket, returns a signed JWT. Passwords are stored as bcrypt hashes. Middleware verifies JWT tokens for protected routes, extracting user ID for ownership checks.
// Rate limiting uses a token bucket algorithm per client IP. Middleware tracks requests in a map with last refill time and tokens. Exceeding limits returns a 429 status. Limits are configurable (default: 100 requests/minute) and exempt internal health checks.
// BoltDB initialization creates buckets if missing. Batch writes improve performance during high traffic. Periodic backup schedules persist the database to cloud storage. In-memory caches are warmed at startup by iterating BoltDB entries.
// Configuration is loaded from environment variables for the HTTP port, JWT secret, BoltDB path, rate limits, and cache TTLs. Fallback defaults are provided for local development. Config validation ensures required fields are set on startup.
// The main function initializes BoltDB connections, caches, and starts the HTTP server with graceful shutdown. Health endpoints (/healthz, /readyz) verify database connectivity and cache population status for load balancers.
// Testing includes unit tests for URL validation, alias generation, and BoltDB CRUD operations. Integration tests use httptest.Server to validate API flows. Benchmark tests measure redirect latency and concurrent create/read throughput.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Config holds the service configuration settings.
type Config struct {
	Port       string
	JWTSecret  string
	BoltDBPath string
	RateLimit  int
	CacheTTL   time.Duration
}

