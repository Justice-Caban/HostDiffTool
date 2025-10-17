package data

import (
	"os"
	"testing"
	time "time"
)

func TestNewDB(t *testing.T) {
	dbPath := "./test_new_db.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	if db.db == nil {
		t.Fatal("Database connection is nil")
	}

	// Verify table exists
	row := db.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='snapshots';")
	var tableName string
	err = row.Scan(&tableName)
	if err != nil {
		t.Fatalf("Failed to verify table creation: %v", err)
	}
	if tableName != "snapshots" {
		t.Fatalf("Expected table 'snapshots', got %s", tableName)
	}
}

func TestInsertAndGetSnapshot(t *testing.T) {
	dbPath := "./test_insert_get.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	ip := "192.168.1.1"
	timestamp := time.Now().Format(time.RFC3339)
	data := []byte("{\"key\":\"value\"}")

	id, err := db.InsertSnapshot(ip, timestamp, data)
	if err != nil {
		t.Fatalf("InsertSnapshot failed: %v", err)
	}

	if id == "" {
		t.Fatal("Expected a non-empty ID, got empty")
	}

	snap, err := db.GetSnapshotByID(id)
	if err != nil {
		t.Fatalf("GetSnapshotByID failed: %v", err)
	}
	if snap == nil {
		t.Fatal("Expected snapshot, got nil")
	}
	if snap.IPAddress != ip {
		t.Errorf("Expected IP %s, got %s", ip, snap.IPAddress)
	}
	if snap.Timestamp != timestamp {
		t.Errorf("Expected Timestamp %s, got %s", timestamp, snap.Timestamp)
	}
	if string(snap.Data) != string(data) {
		t.Errorf("Expected Data %s, got %s", string(data), string(snap.Data))
	}

	snapsByIP, err := db.GetSnapshotsByIP(ip)
	if err != nil {
		t.Fatalf("GetSnapshotsByIP failed: %v", err)
	}
	if len(snapsByIP) != 1 {
		t.Fatalf("Expected 1 snapshot for IP, got %d", len(snapsByIP))
	}
	if snapsByIP[0].ID != id {
		t.Errorf("Expected snapshot ID %s, got %s", id, snapsByIP[0].ID)
	}

	// Test duplicate insertion (should fail due to UNIQUE constraint)
	_, err = db.InsertSnapshot(ip, timestamp, data)
	if err == nil {
		t.Fatal("Expected error on duplicate insert, got nil")
	}
}

func TestGetSnapshotNotFound(t *testing.T) {
	dbPath := "./test_not_found.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	snap, err := db.GetSnapshotByID("999")
	if err != nil {
		t.Fatalf("GetSnapshotByID failed: %v", err)
	}
	if snap != nil {
		t.Fatal("Expected nil snapshot for non-existent ID, got non-nil")
	}
}
