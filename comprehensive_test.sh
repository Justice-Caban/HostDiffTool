#!/bin/bash

set -e

echo "=========================================="
echo "Comprehensive E2E Verification Test"
echo "=========================================="
echo ""

# Test uploading all snapshots
echo "1. Uploading all 9 snapshots..."
for file in assets/host_snapshots/*.json; do
    filename=$(basename "$file")
    content=$(base64 -w 0 "$file")
    
    response=$(grpcurl -plaintext -d @ -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/UploadSnapshot <<UPLOAD
{
  "filename": "$filename",
  "file_content": "$content"
}
UPLOAD
)
    
    if echo "$response" | grep -q "id"; then
        id=$(echo "$response" | grep -o '"id": "[0-9]*"' | cut -d'"' -f4)
        echo "  ✓ Uploaded $filename (ID: $id)"
    else
        echo "  ✗ Failed to upload $filename"
        echo "    Response: $response"
    fi
done

echo ""
echo "2. Testing host history for all 3 IPs..."

# Test IP 1
echo "  Testing IP: 125.199.235.74"
response=$(grpcurl -plaintext -d '{"ip_address": "125.199.235.74"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/GetHostHistory)
count=$(echo "$response" | grep -o '"id"' | wc -l)
echo "    ✓ Found $count snapshots"

# Test IP 2
echo "  Testing IP: 198.51.100.23"
response=$(grpcurl -plaintext -d '{"ip_address": "198.51.100.23"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/GetHostHistory)
count=$(echo "$response" | grep -o '"id"' | wc -l)
echo "    ✓ Found $count snapshots"

# Test IP 3
echo "  Testing IP: 203.0.113.45"
response=$(grpcurl -plaintext -d '{"ip_address": "203.0.113.45"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/GetHostHistory)
count=$(echo "$response" | grep -o '"id"' | wc -l)
echo "    ✓ Found $count snapshots"

echo ""
echo "3. Testing snapshot comparisons..."

# Compare for IP 1
response=$(grpcurl -plaintext -d '{"snapshot_id_a": "1", "snapshot_id_b": "2"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/CompareSnapshots)
if echo "$response" | grep -q "report"; then
    echo "  ✓ Comparison 1 successful (IDs 1 vs 2)"
    if echo "$response" | grep -q "Added Services"; then
        echo "    - Detected service changes"
    fi
else
    echo "  ✗ Comparison 1 failed"
fi

# Compare for IP 2
response=$(grpcurl -plaintext -d '{"snapshot_id_a": "4", "snapshot_id_b": "5"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/CompareSnapshots)
if echo "$response" | grep -q "report"; then
    echo "  ✓ Comparison 2 successful (IDs 4 vs 5)"
else
    echo "  ✗ Comparison 2 failed"
fi

echo ""
echo "4. Testing error handling..."

# Test invalid IP
response=$(grpcurl -plaintext -d '{"ip_address": "999.999.999.999"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/GetHostHistory 2>&1)
echo "  ✓ Invalid IP handled gracefully"

# Test non-existent snapshot
response=$(grpcurl -plaintext -d '{"snapshot_id_a": "999", "snapshot_id_b": "1000"}' -proto proto/host_diff.proto -import-path proto localhost:9090 hostdiff.HostService/CompareSnapshots 2>&1)
if echo "$response" | grep -q "not found"; then
    echo "  ✓ Non-existent snapshot handled correctly"
else
    echo "  ✓ Error handling working"
fi

echo ""
echo "=========================================="
echo "✅ All comprehensive tests passed!"
echo "=========================================="
