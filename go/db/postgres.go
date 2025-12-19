package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// PostgresConnector handles PostgreSQL database operations
type PostgresConnector struct {
	db *sql.DB
}

// NewPostgresConnector creates a new PostgreSQL connector
func NewPostgresConnector(dsn string) (*PostgresConnector, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &PostgresConnector{db: db}, nil
}

// Close closes the database connection
func (p *PostgresConnector) Close() error {
	return p.db.Close()
}

// BatchInsert performs a batch insert using PostgreSQL COPY or multi-value INSERT
func (p *PostgresConnector) BatchInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Use multi-value INSERT for better compatibility
	// COPY would be faster but requires more complex setup
	return p.multiValueInsert(ctx, table, columns, rows)
}

// multiValueInsert performs a multi-value INSERT statement
func (p *PostgresConnector) multiValueInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error {
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
			placeholders[j] = fmt.Sprintf("$%d", len(args)+j+1)
		}
		valuePlaceholders[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
		args = append(args, row...)
	}

	// Build and execute query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, colList, strings.Join(valuePlaceholders, ", "))
	
	_, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	return nil
}

// StreamQuery executes a query and returns rows for streaming
func (p *PostgresConnector) StreamQuery(ctx context.Context, query string) (*sql.Rows, error) {
	// Use a transaction with a cursor for large result sets
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set fetch size for cursor
	if _, err := tx.ExecContext(ctx, "SET cursor_tuple_fraction = 1.0"); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set cursor options: %w", err)
	}

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return rows, nil
}

// GetColumns returns the column names from a query result
func (p *PostgresConnector) GetColumns(rows *sql.Rows) ([]string, error) {
	return rows.Columns()
}
