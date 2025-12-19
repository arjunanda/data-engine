package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Connector is the interface for database operations
type Connector interface {
	BatchInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error
	StreamQuery(ctx context.Context, query string) (*sql.Rows, error)
	GetColumns(rows *sql.Rows) ([]string, error)
	Close() error
}

// NewConnector creates a new database connector based on DSN
func NewConnector(dsn string) (Connector, error) {
	// Detect database type from DSN
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return NewPostgresConnector(dsn)
	}
	
	// MySQL DSN formats: user:pass@tcp(host:port)/db or mysql://...
	if strings.Contains(dsn, "@tcp(") || strings.HasPrefix(dsn, "mysql://") {
		return NewMySQLConnector(dsn)
	}

	return nil, fmt.Errorf("unsupported database type in DSN: %s", dsn)
}
