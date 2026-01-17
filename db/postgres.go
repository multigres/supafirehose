package db

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
)

// ConnectionManager tracks active connections
type ConnectionManager struct {
	connString        string
	activeConnections atomic.Int32
	totalCreated      atomic.Int64
	totalFailed       atomic.Int64
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(connString string) *ConnectionManager {
	return &ConnectionManager{
		connString: connString,
	}
}

// Connect creates a new direct connection to the database
func (cm *ConnectionManager) Connect(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, cm.connString)
	if err != nil {
		cm.totalFailed.Add(1)
		if cm.totalFailed.Load()%100 == 1 {
			log.Printf("Connection failed (total failures: %d): %v", cm.totalFailed.Load(), err)
		}
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	cm.activeConnections.Add(1)
	cm.totalCreated.Add(1)
	if cm.totalCreated.Load()%1000 == 0 {
		log.Printf("Connections: active=%d, total_created=%d, total_failed=%d",
			cm.activeConnections.Load(), cm.totalCreated.Load(), cm.totalFailed.Load())
	}
	return conn, nil
}

// Release decrements the connection counter (call when closing a connection)
func (cm *ConnectionManager) Release() {
	cm.activeConnections.Add(-1)
}

// ActiveConnections returns the current count of active connections
func (cm *ConnectionManager) ActiveConnections() int32 {
	return cm.activeConnections.Load()
}

// Ping verifies connectivity to the database
func (cm *ConnectionManager) Ping(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, cm.connString)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}

// GetDatabaseSize returns the current database size in bytes
func (cm *ConnectionManager) GetDatabaseSize(ctx context.Context) (int64, error) {
	conn, err := pgx.Connect(ctx, cm.connString)
	if err != nil {
		return 0, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close(ctx)

	var size int64
	err = conn.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("failed to get database size: %w", err)
	}
	return size, nil
}
