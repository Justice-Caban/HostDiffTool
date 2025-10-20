package data

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Snapshot represents a host snapshot stored in the database.
type Snapshot struct {
	ID        string
	IPAddress string
	Timestamp string
	Data      []byte
}

// DB provides database access.
type DB struct {
	db *sql.DB
}

// NewDB initializes a new SQLite database connection and creates the necessary table.
// Configures performance pragmas and connection pooling for optimal operation.
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// For SQLite, we typically want a single writer connection
	db.SetMaxOpenConns(1)       // Prevent database locked errors
	db.SetMaxIdleConns(1)       // Keep one connection alive
	db.SetConnMaxLifetime(0)    // Connections never expire

	// Enable performance pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",        // Write-Ahead Logging for better concurrency
		"PRAGMA synchronous=NORMAL;",      // Balance between safety and speed
		"PRAGMA cache_size=-64000;",       // 64MB cache (negative = KB)
		"PRAGMA temp_store=MEMORY;",       // Use memory for temp tables
		"PRAGMA mmap_size=268435456;",     // 256MB memory-mapped I/O
		"PRAGMA busy_timeout=5000;",       // Wait up to 5 seconds on locks
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	// Create the snapshots table if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		data BLOB NOT NULL,
		UNIQUE(ip_address, timestamp)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Create index if it doesn't exist (for faster queries by IP)
	indexSQL := `
	CREATE INDEX IF NOT EXISTS idx_ip_timestamp
	ON snapshots(ip_address, timestamp DESC);
	`
	_, err = db.Exec(indexSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &DB{db: db}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// InsertSnapshot inserts a new snapshot into the database.
func (d *DB) InsertSnapshot(ipAddress, timestamp string, data []byte) (string, error) {
	res, err := d.db.Exec(
		"INSERT INTO snapshots (ip_address, timestamp, data) VALUES (?, ?, ?)",
		ipAddress,
		timestamp,
		string(data),
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert snapshot: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return fmt.Sprintf("%d", id), nil
}

// GetSnapshotsByIP retrieves all snapshots for a given IP address.
func (d *DB) GetSnapshotsByIP(ipAddress string) ([]*Snapshot, error) {
	rows, err := d.db.Query(
		"SELECT id, ip_address, timestamp, data FROM snapshots WHERE ip_address = ? ORDER BY timestamp DESC",
		ipAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshots by IP: %w", err)
	}
	defer rows.Close()

	var snapshots []*Snapshot
	for rows.Next() {
		var s Snapshot
		var dataStr string
		if err := rows.Scan(&s.ID, &s.IPAddress, &s.Timestamp, &dataStr); err != nil {
			return nil, fmt.Errorf("failed to scan snapshot row: %w", err)
		}
		s.Data = []byte(dataStr)
		snapshots = append(snapshots, &s)
	}

	return snapshots, nil
}

// GetSnapshotByID retrieves a single snapshot by its ID.
func (d *DB) GetSnapshotByID(id string) (*Snapshot, error) {
	var s Snapshot
	var dataStr string

	err := d.db.QueryRow(
		"SELECT id, ip_address, timestamp, data FROM snapshots WHERE id = ?",
		id,
	).Scan(&s.ID, &s.IPAddress, &s.Timestamp, &dataStr)

	if err == sql.ErrNoRows {
		return nil, nil // No snapshot found with this ID
	} else if err != nil {
		return nil, fmt.Errorf("failed to query snapshot by ID: %w", err)
	}

	s.Data = []byte(dataStr)
	return &s, nil
}