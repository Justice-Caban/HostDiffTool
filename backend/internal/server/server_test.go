package server

import (
	"testing"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		filename  string
		expectedIP string
		expectedTimestamp string
		expectedErr bool
	}{
		{
			filename:  "host_125.199.235.74_2025-09-10T03-00-00Z.json",
			expectedIP: "125.199.235.74",
			expectedTimestamp: "2025-09-10T03:00:00Z",
			expectedErr: false,
		},
		{
			filename:  "host_198.51.100.23_2025-09-15T08-49-45Z.json",
			expectedIP: "198.51.100.23",
			expectedTimestamp: "2025-09-15T08:49:45Z",
			expectedErr: false,
		},
		{
			filename:  "invalid_filename.json",
			expectedIP: "",
			expectedTimestamp: "",
			expectedErr: true,
		},
		{
			filename:  "host_1.2.3_2025-09-10T03-00-00Z.json", // Invalid IP
			expectedIP: "",
			expectedTimestamp: "",
			expectedErr: true,
		},
		{
			filename:  "host_127.0.0.1_2025-09-10T03-00Z.json", // Invalid timestamp
			expectedIP: "",
			expectedTimestamp: "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		ip, ts, err := parseFilename(tt.filename)

		if (err != nil) != tt.expectedErr {
			t.Errorf("For filename %s, expected error: %v, got: %v", tt.filename, tt.expectedErr, err != nil)
		}

		if !tt.expectedErr {
			if ip != tt.expectedIP {
				t.Errorf("For filename %s, expected IP %s, got %s", tt.filename, tt.expectedIP, ip)
			}
			if ts != tt.expectedTimestamp {
				t.Errorf("For filename %s, expected timestamp %s, got %s", tt.filename, tt.expectedTimestamp, ts)
			}
		}
	}
}
