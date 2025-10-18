package validation

import (
	"testing"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		wantIP        string
		wantTimestamp string
		wantErr       bool
	}{
		{
			name:          "valid filename",
			filename:      "host_127.0.0.1_2025-10-16T12-00-00Z.json",
			wantIP:        "127.0.0.1",
			wantTimestamp: "2025-10-16T12:00:00Z",
			wantErr:       false,
		},
		{
			name:          "valid filename with different IP",
			filename:      "host_192.168.1.100_2025-09-10T03-00-00Z.json",
			wantIP:        "192.168.1.100",
			wantTimestamp: "2025-09-10T03:00:00Z",
			wantErr:       false,
		},
		{
			name:     "invalid format - missing host prefix",
			filename: "snapshot_127.0.0.1_2025-10-16T12-00-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid format - missing extension",
			filename: "host_127.0.0.1_2025-10-16T12-00-00Z",
			wantErr:  true,
		},
		{
			name:     "invalid IP - octet > 255",
			filename: "host_256.0.0.1_2025-10-16T12-00-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid IP - negative octet",
			filename: "host_-1.0.0.1_2025-10-16T12-00-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid timestamp - month 13",
			filename: "host_127.0.0.1_2025-13-16T12-00-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid timestamp - day 32",
			filename: "host_127.0.0.1_2025-10-32T12-00-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid timestamp - hour 24",
			filename: "host_127.0.0.1_2025-10-16T24-00-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid timestamp - minute 60",
			filename: "host_127.0.0.1_2025-10-16T12-60-00Z.json",
			wantErr:  true,
		},
		{
			name:     "invalid timestamp - second 60",
			filename: "host_127.0.0.1_2025-10-16T12-00-60Z.json",
			wantErr:  true,
		},
		{
			name:          "edge case - IP 0.0.0.0",
			filename:      "host_0.0.0.0_2025-10-16T12-00-00Z.json",
			wantIP:        "0.0.0.0",
			wantTimestamp: "2025-10-16T12:00:00Z",
			wantErr:       false,
		},
		{
			name:          "edge case - IP 255.255.255.255",
			filename:      "host_255.255.255.255_2025-10-16T12-00-00Z.json",
			wantIP:        "255.255.255.255",
			wantTimestamp: "2025-10-16T12:00:00Z",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFilename(tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.IPAddress != tt.wantIP {
					t.Errorf("ParseFilename() IPAddress = %v, want %v", result.IPAddress, tt.wantIP)
				}
				if result.Timestamp != tt.wantTimestamp {
					t.Errorf("ParseFilename() Timestamp = %v, want %v", result.Timestamp, tt.wantTimestamp)
				}
			}
		})
	}
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{"valid IP", "192.168.1.1", false},
		{"valid IP all zeros", "0.0.0.0", false},
		{"valid IP all max", "255.255.255.255", false},
		{"invalid octet 256", "256.0.0.1", true},
		{"invalid octet 999", "999.999.999.999", true},
		{"invalid format", "192.168.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPAddress(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIPAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAndNormalizeTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		want      string
		wantErr   bool
	}{
		{
			name:      "valid timestamp",
			timestamp: "2025-10-16T12-00-00Z",
			want:      "2025-10-16T12:00:00Z",
			wantErr:   false,
		},
		{
			name:      "valid timestamp different time",
			timestamp: "2025-01-01T23-59-59Z",
			want:      "2025-01-01T23:59:59Z",
			wantErr:   false,
		},
		{
			name:      "invalid month 0",
			timestamp: "2025-00-16T12-00-00Z",
			wantErr:   true,
		},
		{
			name:      "invalid month 13",
			timestamp: "2025-13-16T12-00-00Z",
			wantErr:   true,
		},
		{
			name:      "invalid day 0",
			timestamp: "2025-10-00T12-00-00Z",
			wantErr:   true,
		},
		{
			name:      "invalid day 32",
			timestamp: "2025-10-32T12-00-00Z",
			wantErr:   true,
		},
		{
			name:      "invalid hour 24",
			timestamp: "2025-10-16T24-00-00Z",
			wantErr:   true,
		},
		{
			name:      "invalid minute 60",
			timestamp: "2025-10-16T12-60-00Z",
			wantErr:   true,
		},
		{
			name:      "invalid second 60",
			timestamp: "2025-10-16T12-00-60Z",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAndNormalizeTimestamp(tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndNormalizeTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("validateAndNormalizeTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}
