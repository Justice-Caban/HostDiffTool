package diff

import (
	"strings"
	"testing"
)

func TestDiffSnapshots_NoDifferences(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"timestamp": "2023-01-01T00:00:00Z",
		"services": [
			{
				"port": 80,
				"protocol": "HTTP",
				"status": 200,
				"software": {"vendor": "apache", "product": "httpd", "version": "2.4.57"}
			}
		],
		"service_count": 1
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotA)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if !strings.Contains(report.Summary, "No meaningful differences found") {
		t.Errorf("Expected no differences summary, got: %s", report.Summary)
	}
	if len(report.AddedServices) > 0 || len(report.RemovedServices) > 0 ||
		len(report.ChangedServices) > 0 || len(report.AddedCVEs) > 0 || len(report.RemovedCVEs) > 0 {
		t.Errorf("Expected empty diff report fields, got non-empty")
	}
}

func TestDiffSnapshots_ServiceAdded(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"timestamp": "2023-01-01T00:00:00Z",
		"services": [
			{"port": 80, "protocol": "HTTP", "status": 200}
		],
		"service_count": 1
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"timestamp": "2023-01-02T00:00:00Z",
		"services": [
			{"port": 80, "protocol": "HTTP", "status": 200},
			{"port": 443, "protocol": "HTTPS", "status": 200}
		],
		"service_count": 2
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.AddedServices) != 1 {
		t.Errorf("Expected 1 added service, got %d", len(report.AddedServices))
	}
	if len(report.AddedServices) > 0 && report.AddedServices[0].Port != 443 {
		t.Errorf("Expected added port 443, got %d", report.AddedServices[0].Port)
	}
}

func TestDiffSnapshots_ServiceRemoved(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP"},
			{"port": 22, "protocol": "SSH"}
		],
		"service_count": 2
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP"}
		],
		"service_count": 1
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.RemovedServices) != 1 {
		t.Errorf("Expected 1 removed service, got %d", len(report.RemovedServices))
	}
	if len(report.RemovedServices) > 0 && report.RemovedServices[0].Port != 22 {
		t.Errorf("Expected removed port 22, got %d", report.RemovedServices[0].Port)
	}
}

func TestDiffSnapshots_ServiceChanged_Status(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP", "status": 200}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 80, "protocol": "HTTP", "status": 301}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Errorf("Expected 1 changed service, got %d", len(report.ChangedServices))
	}
	if len(report.ChangedServices) > 0 {
		if report.ChangedServices[0].Port != 80 {
			t.Errorf("Expected changed port 80, got %d", report.ChangedServices[0].Port)
		}
		if _, ok := report.ChangedServices[0].Changes["status"]; !ok {
			t.Errorf("Expected status change, got changes: %v", report.ChangedServices[0].Changes)
		}
	}
}

func TestDiffSnapshots_ServiceChanged_SoftwareVersion(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 22,
				"protocol": "SSH",
				"software": {"vendor": "openssh", "product": "openssh", "version": "8.2p1"}
			}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 22,
				"protocol": "SSH",
				"software": {"vendor": "openssh", "product": "openssh", "version": "8.4p1"}
			}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Errorf("Expected 1 changed service, got %d", len(report.ChangedServices))
	}
	if len(report.ChangedServices) > 0 {
		if _, ok := report.ChangedServices[0].Changes["software_version"]; !ok {
			t.Errorf("Expected software_version change, got changes: %v", report.ChangedServices[0].Changes)
		}
	}
}

func TestDiffSnapshots_TLSAdded(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{"port": 443, "protocol": "HTTPS"}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 443,
				"protocol": "HTTPS",
				"tls": {"version": "tlsv1_2", "cipher": "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"}
			}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.ChangedServices) != 1 {
		t.Errorf("Expected 1 changed service, got %d", len(report.ChangedServices))
	}
	if len(report.ChangedServices) > 0 {
		if _, ok := report.ChangedServices[0].Changes["tls"]; !ok {
			t.Errorf("Expected TLS change, got changes: %v", report.ChangedServices[0].Changes)
		}
	}
}

func TestDiffSnapshots_CVEAdded(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 22,
				"protocol": "SSH",
				"vulnerabilities": ["CVE-2020-99990"]
			}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 22,
				"protocol": "SSH",
				"vulnerabilities": ["CVE-2020-99990", "CVE-2023-99992"]
			}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.AddedCVEs) != 1 {
		t.Errorf("Expected 1 added CVE, got %d", len(report.AddedCVEs))
	}
	if len(report.AddedCVEs) > 0 && report.AddedCVEs[0].CVEID != "CVE-2023-99992" {
		t.Errorf("Expected CVE-2023-99992, got %s", report.AddedCVEs[0].CVEID)
	}
}

func TestDiffSnapshots_CVERemoved(t *testing.T) {
	snapshotA := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 22,
				"protocol": "SSH",
				"vulnerabilities": ["CVE-2020-99990", "CVE-2023-99992"]
			}
		]
	}`)
	snapshotB := []byte(`{
		"ip": "127.0.0.1",
		"services": [
			{
				"port": 22,
				"protocol": "SSH",
				"vulnerabilities": ["CVE-2020-99990"]
			}
		]
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	if len(report.RemovedCVEs) != 1 {
		t.Errorf("Expected 1 removed CVE, got %d", len(report.RemovedCVEs))
	}
	if len(report.RemovedCVEs) > 0 && report.RemovedCVEs[0].CVEID != "CVE-2023-99992" {
		t.Errorf("Expected CVE-2023-99992, got %s", report.RemovedCVEs[0].CVEID)
	}
}

func TestDiffSnapshots_RealCensysData(t *testing.T) {
	// Test with actual Censys snapshot format
	snapshotA := []byte(`{
		"timestamp": "2025-09-10T03:00:00Z",
		"ip": "125.199.235.74",
		"services": [
			{
				"port": 80,
				"protocol": "HTTP",
				"status": 200,
				"software": {
					"vendor": "microsoft",
					"product": "internet_information_services",
					"version": "8.5"
				}
			}
		],
		"service_count": 1
	}`)

	snapshotB := []byte(`{
		"timestamp": "2025-09-15T08:49:45Z",
		"ip": "125.199.235.74",
		"services": [
			{
				"port": 80,
				"protocol": "HTTP",
				"status": 301,
				"software": {
					"vendor": "microsoft",
					"product": "internet_information_services",
					"version": "8.5"
				}
			},
			{
				"port": 443,
				"protocol": "HTTPS",
				"status": 200,
				"software": {
					"vendor": "microsoft",
					"product": "asp.net"
				},
				"tls": {
					"version": "tlsv1_2",
					"cipher": "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
				},
				"vulnerabilities": ["CVE-2023-99999"]
			}
		],
		"service_count": 2
	}`)

	report, err := DiffSnapshots(snapshotA, snapshotB)
	if err != nil {
		t.Fatalf("DiffSnapshots failed: %v", err)
	}

	// Should have 1 added service (port 443)
	if len(report.AddedServices) != 1 {
		t.Errorf("Expected 1 added service, got %d", len(report.AddedServices))
	}

	// Should have 1 changed service (port 80 status change)
	if len(report.ChangedServices) != 1 {
		t.Errorf("Expected 1 changed service, got %d", len(report.ChangedServices))
	}

	// Should have 1 added CVE
	if len(report.AddedCVEs) != 1 {
		t.Errorf("Expected 1 added CVE, got %d", len(report.AddedCVEs))
	}

	// Verify summary contains expected text
	if !strings.Contains(report.Summary, "Added Services") {
		t.Errorf("Expected summary to contain 'Added Services', got: %s", report.Summary)
	}
	if !strings.Contains(report.Summary, "Changed Services") {
		t.Errorf("Expected summary to contain 'Changed Services', got: %s", report.Summary)
	}
	if !strings.Contains(report.Summary, "Added Vulnerabilities") {
		t.Errorf("Expected summary to contain 'Added Vulnerabilities', got: %s", report.Summary)
	}
}
