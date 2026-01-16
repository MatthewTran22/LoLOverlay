package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Build-time variables - set via -ldflags
// Example: go build -ldflags "-X 'ghostdraft/internal/data.TursoURL=libsql://...' -X 'ghostdraft/internal/data.TursoAuthToken=...'"
var (
	TursoURL       string // Turso database URL
	TursoAuthToken string // Turso auth token (read-only)
)

// TursoClient wraps a connection to Turso with caching
type TursoClient struct {
	db    *sql.DB
	cache *QueryCache
}

// QueryCache provides thread-safe in-memory caching
type QueryCache struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewQueryCache creates a new query cache
func NewQueryCache() *QueryCache {
	return &QueryCache{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value from the cache
func (c *QueryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// Set stores a value in the cache
func (c *QueryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Clear removes all cached values
func (c *QueryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]interface{})
}

// NewTursoClient creates a new Turso client with caching
func NewTursoClient() (*TursoClient, error) {
	url := TursoURL
	token := TursoAuthToken

	// Fall back to environment variables for development
	if url == "" {
		url = getEnv("TURSO_DATABASE_URL", "")
	}
	if token == "" {
		token = getEnv("TURSO_AUTH_TOKEN", "")
	}

	if url == "" {
		return nil, fmt.Errorf("Turso URL not configured (set TURSO_DATABASE_URL or build with -ldflags)")
	}

	connStr := url
	if token != "" {
		connStr = fmt.Sprintf("%s?authToken=%s", url, token)
	}

	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Turso: %w", err)
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping Turso: %w", err)
	}

	fmt.Println("[Turso] Connected successfully")

	return &TursoClient{
		db:    db,
		cache: NewQueryCache(),
	}, nil
}

// Close closes the Turso connection
func (c *TursoClient) Close() error {
	return c.db.Close()
}

// GetDB returns the underlying database connection
func (c *TursoClient) GetDB() *sql.DB {
	return c.db
}

// GetCache returns the query cache
func (c *TursoClient) GetCache() *QueryCache {
	return c.cache
}

// ClearCache clears all cached query results
func (c *TursoClient) ClearCache() {
	c.cache.Clear()
	fmt.Println("[Turso] Cache cleared")
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
