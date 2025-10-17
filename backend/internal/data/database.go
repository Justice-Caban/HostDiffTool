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
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create the snapshots table if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		data TEXT NOT NULL,
		UNIQUE(ip_address, timestamp)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
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