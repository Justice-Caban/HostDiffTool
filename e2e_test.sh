#!/bin/bash

set -e

# --- Test Cases ---

# Test Case 1: Upload a snapshot
SNAPSHOT_FILE_1="./assets/host_snapshots/host_125.199.235.74_2025-09-10T03-00-00Z.json"
FILENAME_1="host_125.199.235.74_2025-09-10T03-00-00Z.json"
IP_ADDRESS_1="125.199.235.74"

FILE_CONTENT_1=$(base64 -w 0 "$SNAPSHOT_FILE_1")

UPLOAD_RESPONSE_1=$(grpcurl -plaintext -d @ -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/UploadSnapshot <<EOF
{
  "filename": "$FILENAME_1",
  "file_content": "$FILE_CONTENT_1"
}
EOF
)

if echo "$UPLOAD_RESPONSE_1" | grep -q "id"; then
  SNAPSHOT_ID_1=$(echo "$UPLOAD_RESPONSE_1" | grep -o '"id": "[0-9]*"' | cut -d'"' -f4)
  echo "Snapshot 1 uploaded. ID: $SNAPSHOT_ID_1"
else
  echo "Failed to upload snapshot 1. Response: $UPLOAD_RESPONSE_1"
  exit 1
fi

# Test Case 2: Upload a second snapshot for the same IP
SNAPSHOT_FILE_2="./assets/host_snapshots/host_125.199.235.74_2025-09-15T08-49-45Z.json"
FILENAME_2="host_125.199.235.74_2025-09-15T08-49-45Z.json"

FILE_CONTENT_2=$(base64 -w 0 "$SNAPSHOT_FILE_2")

UPLOAD_RESPONSE_2=$(grpcurl -plaintext -d @ -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/UploadSnapshot <<EOF
{
  "filename": "$FILENAME_2",
  "file_content": "$FILE_CONTENT_2"
}
EOF
)

if echo "$UPLOAD_RESPONSE_2" | grep -q "id"; then
  SNAPSHOT_ID_2=$(echo "$UPLOAD_RESPONSE_2" | grep -o '"id": "[0-9]*"' | cut -d'"' -f4)
  echo "Snapshot 2 uploaded. ID: $SNAPSHOT_ID_2"
else
  echo "Failed to upload snapshot 2. Response: $UPLOAD_RESPONSE_2"
  exit 1
fi

# Test Case 3: Get host history
HISTORY_RESPONSE=$(grpcurl -plaintext -d @ -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/GetHostHistory <<EOF
{
  "ip_address": "$IP_ADDRESS_1"
}
EOF
)

if echo "$HISTORY_RESPONSE" | grep -q "snapshots"; then
  echo "Host history retrieved successfully."
  echo "Response: $HISTORY_RESPONSE"
else
  echo "Failed to get host history. Response: $HISTORY_RESPONSE"
  exit 1
fi

# Test Case 4: Compare snapshots
COMPARE_RESPONSE=$(grpcurl -plaintext -d @ -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/CompareSnapshots <<EOF
{
  "snapshot_id_a": "$SNAPSHOT_ID_1",
  "snapshot_id_b": "$SNAPSHOT_ID_2"
}
EOF
)

if echo "$COMPARE_RESPONSE" | grep -q "report"; then
  echo "Snapshots compared successfully. Report: $COMPARE_RESPONSE"
else
  echo "Failed to compare snapshots. Response: $COMPARE_RESPONSE"
  exit 1
fi

echo "All end-to-end tests passed!"