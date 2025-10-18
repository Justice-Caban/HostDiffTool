// Package diff contains the core snapshot comparison logic.
package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// HostSnapshot represents the actual Censys host snapshot JSON structure.
type HostSnapshot struct {
	IP           string        `json:"ip"`
	Timestamp    string        `json:"timestamp"`
	Services     []ServiceInfo `json:"services"`
	ServiceCount int           `json:"service_count"`
}

// ServiceInfo represents a service running on a specific port.
type ServiceInfo struct {
	Port            int                    `json:"port"`
	Protocol        string                 `json:"protocol"`
	Status          int                    `json:"status,omitempty"`
	Software        SoftwareInfo           `json:"software,omitempty"`
	TLS             *TLSInfo               `json:"tls,omitempty"`
	Vulnerabilities []string               `json:"vulnerabilities,omitempty"`
	Extra           map[string]interface{} `json:"-"` // For any additional fields
}

// SoftwareInfo represents software details.
type SoftwareInfo struct {
	Vendor  string `json:"vendor,omitempty"`
	Product string `json:"product,omitempty"`
	Version string `json:"version,omitempty"`
}

// TLSInfo represents TLS/SSL configuration details.
type TLSInfo struct {
	Version              string `json:"version,omitempty"`
	Cipher               string `json:"cipher,omitempty"`
	CertFingerprintSHA256 string `json:"cert_fingerprint_sha256,omitempty"`
}

// DiffReport contains the structured differences between two snapshots.
type DiffReport struct {
	Summary         string
	AddedServices   []ServiceInfo
	RemovedServices []ServiceInfo
	ChangedServices []ServiceChange
	AddedCVEs       []CVEChange
	RemovedCVEs     []CVEChange
}

// ServiceChange describes a change in a service's attributes.
type ServiceChange struct {
	Port     int
	Protocol string
	Changes  map[string]string
}

// CVEChange describes a CVE addition or removal.
type CVEChange struct {
	CVEID    string
	Port     int
	Protocol string
}

// DiffSnapshots compares two JSON snapshots and returns a DiffReport.
func DiffSnapshots(snapshotA, snapshotB []byte) (*DiffReport, error) {
	var snapA, snapB HostSnapshot

	if err := json.Unmarshal(snapshotA, &snapA); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot A: %w", err)
	}
	if err := json.Unmarshal(snapshotB, &snapB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot B: %w", err)
	}

	report := &DiffReport{}

	// Compare Services (which include ports)
	compareServices(snapA.Services, snapB.Services, report)

	// Compare Vulnerabilities
	compareVulnerabilities(snapA.Services, snapB.Services, report)

	// Generate a summary string
	report.Summary = generateSummary(report)
	return report, nil
}

func compareServices(servicesA, servicesB []ServiceInfo, report *DiffReport) {
	// Create maps keyed by port+protocol for easier comparison
	// This ensures services on the same port but different protocols are tracked separately
	mapA := make(map[string]ServiceInfo)
	for _, s := range servicesA {
		key := fmt.Sprintf("%d-%s", s.Port, s.Protocol)
		mapA[key] = s
	}
	mapB := make(map[string]ServiceInfo)
	for _, s := range servicesB {
		key := fmt.Sprintf("%d-%s", s.Port, s.Protocol)
		mapB[key] = s
	}

	// Check for removed or changed services
	for key, sA := range mapA {
		if sB, ok := mapB[key]; ok {
			// Service exists in both, check for changes
			changes := make(map[string]string)

			if sA.Protocol != sB.Protocol {
				changes["protocol"] = fmt.Sprintf("%s -> %s", sA.Protocol, sB.Protocol)
			}

			if sA.Status != sB.Status && (sA.Status != 0 || sB.Status != 0) {
				changes["status"] = fmt.Sprintf("%d -> %d", sA.Status, sB.Status)
			}

			if sA.Software.Product != sB.Software.Product {
				changes["software_product"] = fmt.Sprintf("%s -> %s", sA.Software.Product, sB.Software.Product)
			}

			if sA.Software.Version != sB.Software.Version {
				changes["software_version"] = fmt.Sprintf("%s -> %s", sA.Software.Version, sB.Software.Version)
			}

			if sA.Software.Vendor != sB.Software.Vendor {
				changes["software_vendor"] = fmt.Sprintf("%s -> %s", sA.Software.Vendor, sB.Software.Vendor)
			}

			// TLS changes
			if (sA.TLS == nil) != (sB.TLS == nil) {
				if sA.TLS == nil {
					changes["tls"] = "added TLS"
				} else {
					changes["tls"] = "removed TLS"
				}
			} else if sA.TLS != nil && sB.TLS != nil {
				if sA.TLS.Version != sB.TLS.Version {
					changes["tls_version"] = fmt.Sprintf("%s -> %s", sA.TLS.Version, sB.TLS.Version)
				}
				if sA.TLS.Cipher != sB.TLS.Cipher {
					changes["tls_cipher"] = fmt.Sprintf("%s -> %s", sA.TLS.Cipher, sB.TLS.Cipher)
				}
			}

			if len(changes) > 0 {
				report.ChangedServices = append(report.ChangedServices, ServiceChange{
					Port:     sA.Port,
					Protocol: sB.Protocol,
					Changes:  changes,
				})
			}
		} else {
			// Service removed
			report.RemovedServices = append(report.RemovedServices, sA)
		}
	}

	// Check for added services
	for key, sB := range mapB {
		if _, ok := mapA[key]; !ok {
			// Service added
			report.AddedServices = append(report.AddedServices, sB)
		}
	}
}

func compareVulnerabilities(servicesA, servicesB []ServiceInfo, report *DiffReport) {
	// Create maps of CVE+Port combination to track CVEs per service
	// Key format: "CVE-ID:Port" to handle same CVE on different ports
	cveMapA := make(map[string]CVEChange)
	cveMapB := make(map[string]CVEChange)

	for _, s := range servicesA {
		for _, cve := range s.Vulnerabilities {
			key := fmt.Sprintf("%s:%d", cve, s.Port)
			cveMapA[key] = CVEChange{
				CVEID:    cve,
				Port:     s.Port,
				Protocol: s.Protocol,
			}
		}
	}

	for _, s := range servicesB {
		for _, cve := range s.Vulnerabilities {
			key := fmt.Sprintf("%s:%d", cve, s.Port)
			cveMapB[key] = CVEChange{
				CVEID:    cve,
				Port:     s.Port,
				Protocol: s.Protocol,
			}
		}
	}

	// Find removed CVEs
	for key, cveChange := range cveMapA {
		if _, ok := cveMapB[key]; !ok {
			report.RemovedCVEs = append(report.RemovedCVEs, cveChange)
		}
	}

	// Find added CVEs
	for key, cveChange := range cveMapB {
		if _, ok := cveMapA[key]; !ok {
			report.AddedCVEs = append(report.AddedCVEs, cveChange)
		}
	}
}

func generateSummary(report *DiffReport) string {
	var summary bytes.Buffer
	summary.WriteString("Diff Report:\n")

	foundChanges := false

	if len(report.AddedServices) > 0 {
		foundChanges = true
		summary.WriteString(fmt.Sprintf("\n  Added Services (%d):\n", len(report.AddedServices)))
		for _, s := range report.AddedServices {
			summary.WriteString(fmt.Sprintf("    + Port %d (%s)", s.Port, s.Protocol))
			if s.Software.Product != "" {
				summary.WriteString(fmt.Sprintf(" - %s", s.Software.Product))
				if s.Software.Version != "" {
					summary.WriteString(fmt.Sprintf(" %s", s.Software.Version))
				}
			}
			if len(s.Vulnerabilities) > 0 {
				summary.WriteString(fmt.Sprintf(" [%d CVEs]", len(s.Vulnerabilities)))
			}
			summary.WriteString("\n")
		}
	}

	if len(report.RemovedServices) > 0 {
		foundChanges = true
		summary.WriteString(fmt.Sprintf("\n  Removed Services (%d):\n", len(report.RemovedServices)))
		for _, s := range report.RemovedServices {
			summary.WriteString(fmt.Sprintf("    - Port %d (%s)", s.Port, s.Protocol))
			if s.Software.Product != "" {
				summary.WriteString(fmt.Sprintf(" - %s", s.Software.Product))
				if s.Software.Version != "" {
					summary.WriteString(fmt.Sprintf(" %s", s.Software.Version))
				}
			}
			summary.WriteString("\n")
		}
	}

	if len(report.ChangedServices) > 0 {
		foundChanges = true
		summary.WriteString(fmt.Sprintf("\n  Changed Services (%d):\n", len(report.ChangedServices)))
		for _, c := range report.ChangedServices {
			summary.WriteString(fmt.Sprintf("    ~ Port %d (%s):\n", c.Port, c.Protocol))
			for key, change := range c.Changes {
				summary.WriteString(fmt.Sprintf("        %s: %s\n", key, change))
			}
		}
	}

	if len(report.AddedCVEs) > 0 {
		foundChanges = true
		summary.WriteString(fmt.Sprintf("\n  Added Vulnerabilities (%d):\n", len(report.AddedCVEs)))
		for _, cve := range report.AddedCVEs {
			summary.WriteString(fmt.Sprintf("    + %s on port %d (%s)\n", cve.CVEID, cve.Port, cve.Protocol))
		}
	}

	if len(report.RemovedCVEs) > 0 {
		foundChanges = true
		summary.WriteString(fmt.Sprintf("\n  Removed Vulnerabilities (%d):\n", len(report.RemovedCVEs)))
		for _, cve := range report.RemovedCVEs {
			summary.WriteString(fmt.Sprintf("    - %s from port %d (%s)\n", cve.CVEID, cve.Port, cve.Protocol))
		}
	}

	if !foundChanges {
		summary.WriteString("  No meaningful differences found.\n")
	}

	return summary.String()
}
