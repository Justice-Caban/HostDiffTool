package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/justicecaban/host-diff-tool/backend/internal/data"
	"github.com/justicecaban/host-diff-tool/proto"
)

// Edge Case 1: Upload with empty filename
func TestUploadSnapshot_EmptyFilename(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.UploadSnapshotRequest{
		Filename:    "",
		FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
	}

	_, err = server.UploadSnapshot(ctx, req)
	if err == nil {
		t.Error("Expected error for empty filename")
	}
}

// Edge Case 2: Upload with malformed filename
func TestUploadSnapshot_MalformedFilename(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	testCases := []string{
		"not_a_host_file.json",
		"host_invalid_ip.json",
		"host_127.0.0.1.json",
		"host_127.0.0.1_invalid_timestamp.json",
		"host_999.999.999.999_2025-01-01T00-00-00Z.json",
		"host_127.0.0.1_2025-13-01T00-00-00Z.json", // Invalid month
	}

	for _, filename := range testCases {
		req := &proto.UploadSnapshotRequest{
			Filename:    filename,
			FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
		}

		_, err = server.UploadSnapshot(ctx, req)
		if err == nil {
			t.Errorf("Expected error for malformed filename: %s", filename)
		}
	}
}

// Edge Case 3: Upload with empty file content
func TestUploadSnapshot_EmptyContent(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte{},
	}

	_, err = server.UploadSnapshot(ctx, req)
	if err == nil {
		t.Error("Expected error for empty file content")
	}
}

// Edge Case 4: Upload with invalid JSON
func TestUploadSnapshot_InvalidJSON(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{this is not valid json}`),
	}

	_, err = server.UploadSnapshot(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid JSON content")
	}
}

// Edge Case 5: Upload duplicate snapshot
func TestUploadSnapshot_Duplicate(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
	}

	// First upload should succeed
	_, err = server.UploadSnapshot(ctx, req)
	if err != nil {
		t.Fatalf("First upload failed: %v", err)
	}

	// Second upload with same IP and timestamp should fail
	_, err = server.UploadSnapshot(ctx, req)
	if err == nil {
		t.Error("Expected error for duplicate snapshot (same IP + timestamp)")
	}
}

// Edge Case 6: Get history for non-existent IP
func TestGetHostHistory_NonExistentIP(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.GetHostHistoryRequest{
		IpAddress: "192.168.1.1",
	}

	resp, err := server.GetHostHistory(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(resp.Snapshots) != 0 {
		t.Errorf("Expected 0 snapshots for non-existent IP, got %d", len(resp.Snapshots))
	}
}

// Edge Case 7: Get history with empty IP
func TestGetHostHistory_EmptyIP(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.GetHostHistoryRequest{
		IpAddress: "",
	}

	resp, err := server.GetHostHistory(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return empty list, not error
	if len(resp.Snapshots) != 0 {
		t.Errorf("Expected 0 snapshots for empty IP, got %d", len(resp.Snapshots))
	}
}

// Edge Case 8: Get history with invalid IP format
func TestGetHostHistory_InvalidIPFormat(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	testCases := []string{
		"not-an-ip",
		"999.999.999.999",
		"192.168.1",
		"192.168.1.1.1",
		"192.168.-1.1",
	}

	for _, ip := range testCases {
		req := &proto.GetHostHistoryRequest{
			IpAddress: ip,
		}

		resp, err := server.GetHostHistory(ctx, req)
		if err != nil {
			t.Errorf("Should not error for invalid IP format %s, got: %v", ip, err)
		}

		// Should just return empty results
		if len(resp.Snapshots) != 0 {
			t.Errorf("Expected 0 snapshots for invalid IP %s", ip)
		}
	}
}

// Edge Case 9: Compare with non-existent snapshot ID
func TestCompareSnapshots_NonExistentID(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.CompareSnapshotsRequest{
		SnapshotIdA: "999",
		SnapshotIdB: "1000",
	}

	_, err = server.CompareSnapshots(ctx, req)
	if err == nil {
		t.Error("Expected error for non-existent snapshot IDs")
	}
}

// Edge Case 10: Compare with empty snapshot IDs
func TestCompareSnapshots_EmptyIDs(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.CompareSnapshotsRequest{
		SnapshotIdA: "",
		SnapshotIdB: "",
	}

	_, err = server.CompareSnapshots(ctx, req)
	if err == nil {
		t.Error("Expected error for empty snapshot IDs")
	}
}

// Edge Case 11: Compare snapshots from different IPs
func TestCompareSnapshots_DifferentIPs(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	// Upload snapshot for IP 1
	req1 := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
	}
	resp1, err := server.UploadSnapshot(ctx, req1)
	if err != nil {
		t.Fatalf("Failed to upload snapshot 1: %v", err)
	}

	// Upload snapshot for IP 2
	req2 := &proto.UploadSnapshotRequest{
		Filename:    "host_192.168.1.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{"ip": "192.168.1.1", "services": []}`),
	}
	resp2, err := server.UploadSnapshot(ctx, req2)
	if err != nil {
		t.Fatalf("Failed to upload snapshot 2: %v", err)
	}

	// Try to compare snapshots from different IPs
	compareReq := &proto.CompareSnapshotsRequest{
		SnapshotIdA: resp1.Id,
		SnapshotIdB: resp2.Id,
	}

	_, err = server.CompareSnapshots(ctx, compareReq)
	if err == nil {
		t.Error("Expected error when comparing snapshots from different IPs")
	}
}

// Edge Case 12: Compare snapshot with itself
func TestCompareSnapshots_SameSnapshot(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	// Upload a snapshot
	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{"ip": "127.0.0.1", "services": [{"port": 80, "protocol": "HTTP"}]}`),
	}
	resp, err := server.UploadSnapshot(ctx, req)
	if err != nil {
		t.Fatalf("Failed to upload snapshot: %v", err)
	}

	// Compare with itself
	compareReq := &proto.CompareSnapshotsRequest{
		SnapshotIdA: resp.Id,
		SnapshotIdB: resp.Id,
	}

	compareResp, err := server.CompareSnapshots(ctx, compareReq)
	if err != nil {
		t.Fatalf("Unexpected error comparing snapshot with itself: %v", err)
	}

	// Should show no differences
	if compareResp.Report == nil || compareResp.Report.Summary == "" {
		t.Error("Expected valid diff report")
	}
}

// Edge Case 13: Very large snapshot (stress test)
func TestUploadSnapshot_LargeSnapshot(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	// Create a large snapshot with many services
	largeSnapshot := `{"ip": "127.0.0.1", "services": [`
	for i := 0; i < 1000; i++ {
		if i > 0 {
			largeSnapshot += ","
		}
		port := i + 1000
		largeSnapshot += fmt.Sprintf(`{"port": %d, "protocol": "TCP"}`, port)
	}
	largeSnapshot += `]}`

	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(largeSnapshot),
	}

	resp, err := server.UploadSnapshot(ctx, req)
	if err != nil {
		t.Fatalf("Failed to upload large snapshot: %v", err)
	}

	if resp.Id == "" {
		t.Error("Expected valid snapshot ID for large snapshot")
	}
}

// Edge Case 14: Special characters in IP (should fail parsing)
func TestUploadSnapshot_SpecialCharsInFilename(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	testCases := []string{
		"host_127.0.0.1'; DROP TABLE snapshots;--_2025-01-01T00-00-00Z.json",
		"host_127.0.0.1<script>alert('xss')</script>_2025-01-01T00-00-00Z.json",
		"host_../../etc/passwd_2025-01-01T00-00-00Z.json",
	}

	for _, filename := range testCases {
		req := &proto.UploadSnapshotRequest{
			Filename:    filename,
			FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
		}

		_, err = server.UploadSnapshot(ctx, req)
		if err == nil {
			t.Errorf("Expected error for malicious filename: %s", filename)
		}
	}
}

// Edge Case 15: Nil context (should handle gracefully)
func TestServer_NilContext(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)

	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
	}

	// This should not panic
	_, err = server.UploadSnapshot(nil, req)
	// May error or succeed depending on implementation, but shouldn't panic
}

// Edge Case 16: Concurrent uploads of same snapshot
func TestUploadSnapshot_Concurrent(t *testing.T) {
	db, err := data.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)
	ctx := context.Background()

	req := &proto.UploadSnapshotRequest{
		Filename:    "host_127.0.0.1_2025-01-01T00-00-00Z.json",
		FileContent: []byte(`{"ip": "127.0.0.1", "services": []}`),
	}

	// Try to upload the same snapshot concurrently
	errorCount := 0
	successCount := 0

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := server.UploadSnapshot(ctx, req)
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Exactly one should succeed, others should fail due to UNIQUE constraint
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful upload in concurrent scenario, got %d", successCount)
	}

	if errorCount != 9 {
		t.Errorf("Expected 9 failed uploads in concurrent scenario, got %d", errorCount)
	}
}
