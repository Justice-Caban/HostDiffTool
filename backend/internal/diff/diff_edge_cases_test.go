package diff

import (
	"fmt"
	"strings"
	"testing"
)

// Edge Case 1: Empty snapshots
func TestDiffSnapshots_BothEmpty(t *testing.T) {
	snapshotA := []byte(`{"ip": "127.0.0.1", "services": [], "service_count": 0}`)
	snapshotB := []byte(`{"ip": "127.0.0.1", "services": [], "service_count": 0}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Expected no error for empty snapshots, got: %v", err)
	}

	if !strings.Contains(report.Summary, "No meaningful differences found") {
		t.Errorf("Expected no differences for empty snapshots")
	}
}

// Edge Case 2: One empty, one with services
func TestDiffSnapshots_OneEmpty(t *testing.T) {
	snapshotA := []byte(`{"ip": "127.0.0.1", "services": [], "service_count": 0}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP"}
		],
		"service_count": 1
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(report.AddedServices) != 1 {
		t.Errorf("Expected 1 added service when going from empty to 1 service, got %d", len(report.AddedServices))
	}

	if len(report.RemovedServices) != 0 {
		t.Errorf("Expected 0 removed services, got %d", len(report.RemovedServices))
	}
}

// Edge Case 3: Invalid JSON
func TestDiffSnapshots_InvalidJSON_A(t *testing.T) {
	snapshotA := []byte(`{invalid json}`)
	snapshotB := []byte(`{"ip": "127.0.0.1", "services": []}`)

	_, err := DiffSnapshots(snapshotA, snapshotB)
	if err == nil {
		t.Error("Expected error for invalid JSON in snapshot A")
	}
}

func TestDiffSnapshots_InvalidJSON_B(t *testing.T) {
	snapshotA := []byte(`{"ip": "127.0.0.1", "services": []}`)
	snapshotB := []byte(`{invalid json}`)

	_, err := DiffSnapshots(snapshotA, snapshotB)
	if err == nil {
		t.Error("Expected error for invalid JSON in snapshot B")
	}
}

// Edge Case 4: Missing required fields
func TestDiffSnapshots_MissingIP(t *testing.T) {
	snapshotA := []byte(`{"services": []}`)
	snapshotB := []byte(`{"services": []}`)

	// Should not error - missing fields just default to zero values
	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error for missing IP field: %v", err)
	}

	if !strings.Contains(report.Summary, "No meaningful differences found") {
		t.Error("Expected no differences")
	}
}

// Edge Case 5: Duplicate ports in single snapshot
func TestDiffSnapshots_DuplicatePorts(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP"},
			{"port": 80, "protocol": "HTTPS"}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP"}
		]
	}`)

	// The last service with port 80 wins (map overwrites)
	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// May detect a change depending on which service wins in the map
	// This is acceptable behavior for malformed data
	if len(report.AddedServices) == 0 && len(report.RemovedServices) == 0 && len(report.ChangedServices) == 0 {
		t.Error("Expected some difference detected for duplicate ports scenario")
	}
}

// Edge Case 6: Very large port number
func TestDiffSnapshots_LargePortNumber(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 65535, "protocol": "TCP"}]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 65536, "protocol": "TCP"}]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.AddedServices) != 1 || len(report.RemovedServices) != 1 {
		t.Error("Expected port change to be detected as remove + add")
	}
}

// Edge Case 7: Empty vulnerabilities array vs nil
func TestDiffSnapshots_EmptyVsNilVulnerabilities(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "vulnerabilities": []}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH"}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Empty array vs nil/missing should be treated as the same
	if len(report.AddedCVEs) != 0 || len(report.RemovedCVEs) != 0 {
		t.Error("Expected no CVE changes between empty array and nil")
	}
}

// Edge Case 8: Same CVE on multiple ports
func TestDiffSnapshots_SameCVEMultiplePorts(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "vulnerabilities": ["CVE-2023-1234"]},
			{"port": 80, "protocol": "HTTP", "vulnerabilities": ["CVE-2023-1234"]}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "vulnerabilities": []},
			{"port": 80, "protocol": "HTTP", "vulnerabilities": ["CVE-2023-1234"]}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should show CVE removed from port 22
	// Note: Current implementation tracks CVE by port, so same CVE on different ports
	// is tracked separately - this is correct behavior
	if len(report.RemovedCVEs) != 1 {
		t.Errorf("Expected 1 removed CVE (from port 22), got %d", len(report.RemovedCVEs))
	}
}

// Edge Case 9: Protocol change on same port
func TestDiffSnapshots_ProtocolChange(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 80, "protocol": "HTTP"}]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 80, "protocol": "HTTPS"}]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Errorf("Expected 1 changed service for protocol change, got %d", len(report.ChangedServices))
	}

	if _, ok := report.ChangedServices[0].Changes["protocol"]; !ok {
		t.Error("Expected protocol change to be detected")
	}
}

// Edge Case 10: Status 0 vs missing status
func TestDiffSnapshots_StatusZeroVsMissing(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 80, "protocol": "HTTP", "status": 0}]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 80, "protocol": "HTTP"}]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Status 0 is treated as "missing" - no change should be reported
	if len(report.ChangedServices) != 0 {
		t.Error("Expected no changes between status 0 and missing status")
	}
}

// Edge Case 11: Very long software version string
func TestDiffSnapshots_LongVersionString(t *testing.T) {
	longVersion := strings.Repeat("1.2.3.", 100) + "final"
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "software": {"version": "8.2p1"}}
		]
	}`)
	snapshotB := `{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "software": {"version": "` + longVersion + `"}}
		]
	}`

	report, err := DiffSnapshots(snapshotA, []byte(snapshotB))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Error("Expected version change to be detected")
	}
}

// Edge Case 12: Unicode and special characters
func TestDiffSnapshots_UnicodeInFields(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP", "software": {"product": "nginx", "version": "1.0"}}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP", "software": {"product": "nginx™", "version": "1.0-β"}}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error with unicode: %v", err)
	}

	if len(report.ChangedServices) == 0 {
		t.Error("Expected changes to be detected for unicode differences")
	}
}

// Edge Case 13: TLS removed
func TestDiffSnapshots_TLSRemoved(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 443,
				"protocol": "HTTPS",
				"tls": {"version": "tlsv1_2"}
			}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 443, "protocol": "HTTPS"}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Errorf("Expected 1 changed service for TLS removal, got %d", len(report.ChangedServices))
	}

	if _, ok := report.ChangedServices[0].Changes["tls"]; !ok {
		t.Error("Expected TLS removal to be detected")
	}
}

// Edge Case 14: Multiple simultaneous changes on same port
func TestDiffSnapshots_MultipleChanges(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 443,
				"protocol": "HTTPS",
				"status": 200,
				"software": {"product": "nginx", "version": "1.20.1"},
				"tls": {"version": "tlsv1_2", "cipher": "AES128"}
			}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 443,
				"protocol": "HTTPS",
				"status": 301,
				"software": {"product": "nginx", "version": "1.22.0"},
				"tls": {"version": "tlsv1_3", "cipher": "AES256"}
			}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Fatalf("Expected 1 changed service, got %d", len(report.ChangedServices))
	}

	changes := report.ChangedServices[0].Changes
	expectedChanges := []string{"status", "software_version", "tls_version", "tls_cipher"}

	for _, expected := range expectedChanges {
		if _, ok := changes[expected]; !ok {
			t.Errorf("Expected change '%s' to be detected, got changes: %v", expected, changes)
		}
	}
}

// Edge Case 15: Completely different snapshots (nothing in common)
func TestDiffSnapshots_CompletelyDifferent(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH"},
			{"port": 80, "protocol": "HTTP"}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 443, "protocol": "HTTPS"},
			{"port": 3306, "protocol": "MySQL"}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.AddedServices) != 2 {
		t.Errorf("Expected 2 added services, got %d", len(report.AddedServices))
	}

	if len(report.RemovedServices) != 2 {
		t.Errorf("Expected 2 removed services, got %d", len(report.RemovedServices))
	}

	if len(report.ChangedServices) != 0 {
		t.Errorf("Expected 0 changed services, got %d", len(report.ChangedServices))
	}
}

// Edge Case 16: Empty string values
func TestDiffSnapshots_EmptyStringValues(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "", "software": {"vendor": "", "product": "", "version": ""}}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP", "software": {"vendor": "nginx", "product": "nginx", "version": "1.0"}}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Error("Expected changes from empty strings to values")
	}
}

// Edge Case 17: Null/missing software object
func TestDiffSnapshots_MissingSoftware(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [{"port": 80, "protocol": "HTTP"}]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP", "software": {"product": "nginx", "version": "1.0"}}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Software added should be detected as a change
	if len(report.ChangedServices) == 0 {
		t.Error("Expected software addition to be detected as change")
	}
}

// Edge Case 18: Case sensitivity in CVE IDs
func TestDiffSnapshots_CVECaseSensitivity(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "vulnerabilities": ["CVE-2023-1234"]}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 22, "protocol": "SSH", "vulnerabilities": ["cve-2023-1234"]}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should detect as different CVEs (case sensitive)
	if len(report.AddedCVEs) != 1 || len(report.RemovedCVEs) != 1 {
		t.Error("Expected case-sensitive CVE comparison")
	}
}

// Edge Case 19: Very large number of services
func TestDiffSnapshots_ManyServices(t *testing.T) {
	// Build snapshot with 100 services, removing one in snapshot B
	servicesA := `{"ip": "127.0.0.1", "services": [`
	servicesB := `{"ip": "127.0.0.1", "services": [`

	firstB := true
	for i := 1; i <= 100; i++ {
		if i > 1 {
			servicesA += ","
		}
		port := i + 1000
		servicesA += `{"port": ` + fmt.Sprintf("%d", port) + `, "protocol": "TCP"}`

		if i != 50 {
			if !firstB {
				servicesB += ","
			}
			firstB = false
			servicesB += `{"port": ` + fmt.Sprintf("%d", port) + `, "protocol": "TCP"}`
		}
	}
	servicesA += `]}`
	servicesB += `]}`

	report, err := DiffSnapshots([]byte(servicesA), []byte(servicesB))
	if err != nil {
		t.Fatalf("Unexpected error with many services: %v", err)
	}

	if len(report.RemovedServices) != 1 {
		t.Errorf("Expected 1 removed service in large dataset, got %d", len(report.RemovedServices))
	}
}

// Edge Case 20: Whitespace differences in JSON (should not affect comparison)
func TestDiffSnapshots_WhitespaceDifferences(t *testing.T) {
	snapshotA := []byte(`{"ip":"127.0.0.1","services":[{"port":80,"protocol":"HTTP"}]}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 80,
				"protocol": "HTTP"
			}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(report.Summary, "No meaningful differences found") {
		t.Error("Expected no differences despite whitespace variations in JSON")
	}
}
