package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnector handles MySQL database operations
type MySQLConnector struct {
	db *sql.DB
}

// NewMySQLConnector creates a new MySQL connector
func NewMySQLConnector(dsn string) (*MySQLConnector, error) {
	// Normalize DSN
	// 1. Handle mysql:// scheme
	if strings.HasPrefix(dsn, "mysql://") {
		u, err := url.Parse(dsn)
		if err == nil {
			userInfo := ""
			if u.User != nil {
				userInfo = u.User.String() + "@"
			}
			path := strings.TrimPrefix(u.Path, "/")
			dsn = fmt.Sprintf("%stcp(%s)/%s", userInfo, u.Host, path)
			if u.RawQuery != "" {
				dsn += "?" + u.RawQuery
			}
		}
	} else if strings.Contains(dsn, "@") && strings.Contains(dsn, "/") && !strings.Contains(dsn, "(") {
		// 2. Handle simplified format: user:pass@host:port/db
		// If it has @ and / but no ( ), assume TCP
		parts := strings.SplitN(dsn, "@", 2)
		rest := parts[1]
		pathParts := strings.SplitN(rest, "/", 2)
		if len(pathParts) == 2 {
			addr := pathParts[0]
			db := pathParts[1]
			dsn = fmt.Sprintf("%s@tcp(%s)/%s", parts[0], addr, db)
		}
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &MySQLConnector{db: db}, nil
}

// Close closes the database connection
func (m *MySQLConnector) Close() error {
	return m.db.Close()
}

// BatchInsert performs a batch insert using multi-value INSERT
func (m *MySQLConnector) BatchInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Build column list
	colList := strings.Join(columns, ", ")

	// Build value placeholders
	valuePlaceholders := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*len(columns))
	
	for i, row := range rows {
		placeholders := make([]string, len(columns))
		for j := range columns {
			placeholders[j] = "?"
		}
		valuePlaceholders[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
		args = append(args, row...)
	}

	// Build and execute query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, colList, strings.Join(valuePlaceholders, ", "))
	
	_, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	return nil
}

// StreamQuery executes a query and returns rows for streaming
func (m *MySQLConnector) StreamQuery(ctx context.Context, query string) (*sql.Rows, error) {
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return rows, nil
}

// GetColumns returns the column names from a query result
func (m *MySQLConnector) GetColumns(rows *sql.Rows) ([]string, error) {
	return rows.Columns()
}
