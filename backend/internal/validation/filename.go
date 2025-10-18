// Package validation provides input validation utilities for the host diff tool.
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// filenamePattern matches the expected filename format: host_<ip>_<timestamp>.json
var filenamePattern = regexp.MustCompile(`host_([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})_([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}Z)\.json`)

// ParsedFilename contains the extracted metadata from a snapshot filename.
type ParsedFilename struct {
	IPAddress string
	Timestamp string // ISO-8601 format
}

// ParseFilename extracts IP address and timestamp from a filename like
// "host_127.0.0.1_2025-10-16T12-00-00Z.json" and validates both components.
//
// Returns an error if:
//   - The filename doesn't match the expected format
//   - The IP address octets are outside the 0-255 range
//   - The timestamp components are invalid (e.g., month > 12)
func ParseFilename(filename string) (*ParsedFilename, error) {
	matches := filenamePattern.FindStringSubmatch(filename)
	if len(matches) != 3 {
		return nil, fmt.Errorf("filename does not match expected format 'host_<ip>_<timestamp>.json': %s", filename)
	}

	ipAddress := matches[1]
	timestampStr := matches[2]

	// Validate IP address
	if err := validateIPAddress(ipAddress); err != nil {
		return nil, err
	}

	// Validate and normalize timestamp
	normalizedTimestamp, err := validateAndNormalizeTimestamp(timestampStr)
	if err != nil {
		return nil, err
	}

	return &ParsedFilename{
		IPAddress: ipAddress,
		Timestamp: normalizedTimestamp,
	}, nil
}

// validateIPAddress checks that each octet is in the valid range 0-255.
func validateIPAddress(ip string) error {
	var octets [4]int
	if _, err := fmt.Sscanf(ip, "%d.%d.%d.%d", &octets[0], &octets[1], &octets[2], &octets[3]); err != nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}

	for i, octet := range octets {
		if octet < 0 || octet > 255 {
			return fmt.Errorf("invalid IP address octet [%d]: %d (must be 0-255)", i, octet)
		}
	}

	return nil
}

// validateAndNormalizeTimestamp validates date/time components and converts
// the timestamp from "2025-10-16T12-00-00Z" to ISO-8601 format "2025-10-16T12:00:00Z".
func validateAndNormalizeTimestamp(timestampStr string) (string, error) {
	var year, month, day, hour, minute, second int
	if _, err := fmt.Sscanf(timestampStr, "%04d-%02d-%02dT%02d-%02d-%02dZ",
		&year, &month, &day, &hour, &minute, &second); err != nil {
		return "", fmt.Errorf("invalid timestamp format: %s", timestampStr)
	}

	// Validate date/time ranges
	if month < 1 || month > 12 {
		return "", fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}
	if day < 1 || day > 31 {
		return "", fmt.Errorf("invalid day: %d (must be 1-31)", day)
	}
	if hour < 0 || hour > 23 {
		return "", fmt.Errorf("invalid hour: %d (must be 0-23)", hour)
	}
	if minute < 0 || minute > 59 {
		return "", fmt.Errorf("invalid minute: %d (must be 0-59)", minute)
	}
	if second < 0 || second > 59 {
		return "", fmt.Errorf("invalid second: %d (must be 0-59)", second)
	}

	// Convert dashes in time portion to colons for ISO-8601 compliance
	// "2025-10-16T12-00-00Z" -> "2025-10-16T12:00:00Z"
	// Only replace dashes after the 'T'
	parts := strings.Split(timestampStr, "T")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid timestamp format: missing T separator")
	}
	timePart := strings.ReplaceAll(parts[1], "-", ":")
	normalized := parts[0] + "T" + timePart
	return normalized, nil
}
