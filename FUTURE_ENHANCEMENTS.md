# Future Enhancements and Known Issues

This document tracks known bugs, limitations, and potential enhancements for the Host Diff Tool.

**Document Version:** 1.0
**Last Updated:** October 2025
**Project Version:** 1.0.0

---

## Table of Contents

- [Critical Issues](#critical-issues)
- [Known Bugs](#known-bugs)
- [Limitations](#limitations)
- [Planned Enhancements](#planned-enhancements)
- [Security Enhancements](#security-enhancements)
- [Performance Optimizations](#performance-optimizations)
- [Developer Experience Improvements](#developer-experience-improvements)

---

## Critical Issues

### 1. Filename Regex Missing Anchors

**Severity:** Medium
**Impact:** Security/Validation

**Description:**
The filename validation regex does not use anchors (`^` and `$`), which means it matches patterns embedded within larger strings.

**Current Code (validation/filename.go:11):**
```go
var filenamePattern = regexp.MustCompile(`host_([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})_([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}Z)\.json`)
```

**Problem:**
- Filename `"prefix_host_127.0.0.1_2025-10-16T12-00-00Z.json"` would PASS validation
- Filename `"host_127.0.0.1_2025-10-16T12-00-00Z.json.backup"` would PASS validation
- Path traversal attempts like `"../../host_127.0.0.1_2025-10-16T12-00-00Z.json"` would PASS

**Recommended Fix:**
```go
var filenamePattern = regexp.MustCompile(`^host_([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})_([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}Z)\.json$`)
```

**Workaround:**
Strip directory paths from filename before validation using `filepath.Base()`.

**References:**
- File: `backend/internal/validation/filename.go:11`
- Related: Path traversal protection

---

### 2. No Validation of Filename Metadata vs JSON Content

**Severity:** Medium
**Impact:** Data Integrity

**Description:**
The system extracts IP address and timestamp from the filename for database storage but does not verify these match the actual JSON content.

**Example Problematic Scenario:**
```bash
# Filename says: 192.168.1.1 at 2025-10-16
# JSON says: 10.0.0.5 at 2025-09-01
# System stores: 192.168.1.1, 2025-10-16 (from filename)
# Diff compares: 10.0.0.5, 2025-09-01 (from JSON)
```

**Impact:**
- Database queries by IP return incorrect results
- Historical tracking becomes unreliable
- Users cannot trust the system metadata

**Current Behavior:**
- `server.go:41` stores `parsed.IPAddress` and `parsed.Timestamp` (from filename)
- `diff.go:68-76` unmarshals JSON and uses `snapA.IP` and `snapA.Timestamp` (from JSON)
- No validation that these match

**Recommended Fix:**
Add validation in `UploadSnapshot`:
```go
// After unmarshaling JSON for validation
var snapshot diff.HostSnapshot
if err := json.Unmarshal(req.GetFileContent(), &snapshot); err != nil {
    return nil, fmt.Errorf("invalid JSON content: %w", err)
}

// Validate metadata matches
if snapshot.IP != parsed.IPAddress {
    return nil, fmt.Errorf("IP mismatch: filename has %s but JSON has %s",
        parsed.IPAddress, snapshot.IP)
}
// Similar check for timestamp
```

**References:**
- File: `backend/internal/server/server.go:27-52`
- Related: Data integrity, user trust

---

### 3. Missing Port in Service Causes Silent Data Loss

**Severity:** Medium
**Impact:** Data Loss

**Description:**
If multiple services in a snapshot have missing or zero `port` values, they all map to the same key `"0-<protocol>"`, and only the last one is retained.

**Example:**
```json
{
  "ip": "127.0.0.1",
  "services": [
    {"protocol": "HTTP", "status": 200},
    {"protocol": "HTTPS", "status": 301},
    {"protocol": "SSH", "status": 22}
  ]
}
```

All three services have missing `port`, so they all get `port: 0`. The service map uses key `"0-HTTP"`, `"0-HTTPS"`, `"0-SSH"`, but if protocols were the same, later services would overwrite earlier ones.

**Current Behavior:**
- No error is raised
- Data is silently lost
- Comparison produces incorrect results

**Recommended Fix:**
Add validation when building service map in `diff.go:91-103`:
```go
for _, s := range servicesA {
    if s.Port == 0 {
        return nil, fmt.Errorf("service missing required field 'port': %+v", s)
    }
    if s.Protocol == "" {
        return nil, fmt.Errorf("service missing required field 'protocol' for port %d", s.Port)
    }
    key := fmt.Sprintf("%d-%s", s.Port, s.Protocol)
    if _, exists := mapA[key]; exists {
        return nil, fmt.Errorf("duplicate service: port %d protocol %s", s.Port, s.Protocol)
    }
    mapA[key] = s
}
```

**Alternative:**
Accept the data but log a warning and generate a unique key for zero-port services.

**References:**
- File: `backend/internal/diff/diff.go:91-103`
- Test: `backend/internal/diff/diff_edge_cases_test.go:87-113`

---

## Known Bugs

### 4. Duplicate Services in Single Snapshot Silently Overwrites

**Severity:** Low
**Impact:** Unexpected Behavior

**Description:**
If a single snapshot contains multiple services with the same `(port, protocol)` combination, the last one in the array silently overwrites previous ones.

**Example:**
```json
{
  "services": [
    {"port": 80, "protocol": "HTTP", "status": 200},
    {"port": 80, "protocol": "HTTP", "status": 301}
  ]
}
```

Only the second service (status 301) is retained.

**Current Behavior:**
- Test `TestDiffSnapshots_DuplicatePorts` documents this as "acceptable behavior for malformed data"
- No error or warning is raised

**Recommended Fix:**
See fix in issue #3 above - validate for duplicates when building the map.

**References:**
- File: `backend/internal/diff/diff.go:91-103`
- Test: `backend/internal/diff/diff_edge_cases_test.go:87-113`

---

### 5. No Pagination in GetHostHistory

**Severity:** Low (becomes High at scale)
**Impact:** Performance, Memory

**Description:**
The `GetHostHistory` endpoint returns ALL snapshots for an IP address with no pagination, filtering, or limits.

**Problem:**
- Host with 10,000 snapshots → response contains 10,000 records
- High memory usage on server and client
- Slow response times
- Poor user experience in frontend

**Current Implementation:**
```go
func (s *Server) GetHostHistory(ctx context.Context, req *proto.GetHostHistoryRequest)
    (*proto.GetHostHistoryResponse, error) {
    snapshots, err := s.db.GetSnapshotsByIP(req.GetIpAddress())
    // Returns everything
}
```

**Recommended Enhancement:**
Add pagination parameters to proto definition:
```protobuf
message GetHostHistoryRequest {
  string ip_address = 1;
  int32 page_size = 2;    // default 50, max 1000
  int32 page = 3;          // default 1
  string order_by = 4;     // "timestamp_desc" (default) or "timestamp_asc"
}

message GetHostHistoryResponse {
  repeated SnapshotInfo snapshots = 1;
  int32 total_count = 2;
  int32 page = 3;
  int32 page_size = 4;
  bool has_next_page = 5;
}
```

**References:**
- File: `backend/internal/server/server.go:54-74`
- Proto: `proto/host_diff.proto`

---

### 6. Empty Protocol vs Non-Empty Protocol Treated as Different Services

**Severity:** Low
**Impact:** Unexpected Behavior

**Description:**
A service with `"protocol": ""` is treated as different from `"protocol": "HTTP"` even if on the same port.

**Example:**
Snapshot A:
```json
{"port": 80, "protocol": ""}
```

Snapshot B:
```json
{"port": 80, "protocol": "HTTP"}
```

**Result:** Reported as one service removed and one added (not a modification).

**Current Behavior:**
- Map keys are `"80-"` vs `"80-HTTP"` → different services
- Test `TestDiffSnapshots_EmptyStringValues` confirms this

**Is This a Bug?**
Debatable. Could be considered correct behavior (empty protocol means "unknown" vs "HTTP").

**Recommended Fix:**
Either:
1. Normalize empty strings to a default value like `"UNKNOWN"` during parsing
2. Document this behavior clearly (already done in assumptions)
3. Reject services with empty protocols (validate required fields)

**References:**
- File: `backend/internal/diff/diff.go:96, 101`
- Test: `backend/internal/diff/diff_edge_cases_test.go` (protocol change tests)

---

### 7. Port Numbers Not Validated (Allows > 65535)

**Severity:** Low
**Impact:** Data Integrity

**Description:**
Port numbers are not validated against the valid range (1-65535). Invalid port numbers like 0, -1, or 99999 are accepted.

**Example:**
```json
{"port": 99999, "protocol": "TCP"}
```

This would be stored and compared without error.

**Current Behavior:**
- `port` is defined as `int` in Go structs
- JSON unmarshaling accepts any integer
- No validation in upload or diff logic

**Recommended Fix:**
Add validation in `UploadSnapshot` after unmarshaling:
```go
for _, service := range snapshot.Services {
    if service.Port < 1 || service.Port > 65535 {
        return nil, fmt.Errorf("invalid port number %d (must be 1-65535)", service.Port)
    }
}
```

**References:**
- File: `backend/internal/diff/diff.go:19-27`
- Test: `backend/internal/diff/diff_edge_cases_test.go:116-134` (accepts 65536)

---

### 8. Timestamp Validation Doesn't Check Calendar Validity

**Severity:** Low
**Impact:** Data Quality

**Description:**
The timestamp validation checks numeric ranges but doesn't validate calendar rules (e.g., February 30, April 31).

**Current Validation:**
```go
if month < 1 || month > 12 { ... }
if day < 1 || day > 31 { ... }
```

**Problem:**
These timestamps would PASS validation:
- `2025-02-30` (February doesn't have 30 days)
- `2025-04-31` (April doesn't have 31 days)
- `2025-02-29` in non-leap years

**Recommended Fix:**
Use Go's `time.Parse()` to validate calendar validity:
```go
func validateAndNormalizeTimestamp(timestampStr string) (string, error) {
    // ... existing code ...

    // Validate calendar correctness
    normalized := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
        year, month, day, hour, minute, second)
    _, err := time.Parse(time.RFC3339, normalized)
    if err != nil {
        return "", fmt.Errorf("invalid date/time: %w", err)
    }

    return normalized, nil
}
```

**References:**
- File: `backend/internal/validation/filename.go:68-104`

---

### 9. SQLite Single-Writer Limitation

**Severity:** Low (becomes High at scale)
**Impact:** Scalability

**Description:**
SQLite allows only one concurrent writer. Under high write load, requests will be serialized or timeout.

**Current Mitigation:**
- WAL mode enabled for better concurrency
- `PRAGMA busy_timeout=5000` allows 5-second wait
- Connection pool limited to 1 connection

**Problem at Scale:**
- 100 concurrent upload requests → serialized execution
- High latency under load
- `database locked` errors if timeout exceeded

**Recommended Enhancement:**
For production at scale, migrate to PostgreSQL or MySQL:
- Support true concurrent writes
- Better performance for large datasets
- Advanced features (replication, partitioning)

**Migration Path:**
1. Abstract database interface in `data/database.go`
2. Implement PostgreSQL adapter
3. Add database driver as config option
4. Update Docker Compose with Postgres container

**References:**
- File: `backend/internal/data/database.go:19-47`
- Documentation: SQLite limitations

---

### 10. Frontend State Lost on Page Refresh

**Severity:** Low
**Impact:** User Experience

**Description:**
All application state (selected snapshots, comparison results, IP address input) is client-side only and lost on browser refresh.

**Problem:**
User scenario:
1. Load history for 192.168.1.1
2. Select two snapshots
3. View diff report
4. Accidentally refresh page
5. All state lost, must start over

**Current Behavior:**
- No state persistence
- No URL routing
- No deep linking

**Recommended Enhancement:**
Implement URL-based routing:
```
http://localhost/history?ip=192.168.1.1
http://localhost/compare?a=123&b=456
```

Or add browser localStorage:
```javascript
useEffect(() => {
  localStorage.setItem('lastIpAddress', ipAddress);
}, [ipAddress]);
```

**References:**
- File: `frontend/src/App.tsx`
- Related: UX improvements

---

## Limitations

### 11. No IPv6 Support

**Severity:** Low (by design)
**Impact:** Feature Gap

**Description:**
Only IPv4 addresses are supported. IPv6 addresses are rejected by filename validation.

**Design Decision:**
Intentionally omitted for simplicity in v1.0.

**Future Enhancement:**
Add IPv6 support:
1. Update filename regex to support IPv6 format
2. Normalize IPv6 addresses (e.g., `::1` vs `0:0:0:0:0:0:0:1`)
3. Update validation logic
4. Add IPv6 test cases
5. Consider dual-stack scenarios

**Filename Format Options:**
- `host_2001-0db8-85a3-0000-0000-8a2e-0370-7334_2025-10-16T12-00-00Z.json`
- `host_[2001:db8:85a3::8a2e:370:7334]_2025-10-16T12-00-00Z.json`

**References:**
- File: `backend/internal/validation/filename.go:11`
- Documentation: Design assumptions in README

---

### 12. No Authentication/Authorization

**Severity:** Low (by design)
**Impact:** Security

**Description:**
The system has no built-in authentication or authorization. Anyone with network access can:
- Upload snapshots
- View all host history
- Compare any snapshots
- Access the entire database

**Design Decision:**
Intentionally omitted for development/demo purposes.

**Security Impact:**
- Suitable only for trusted networks
- Not production-ready without external auth

**Recommended Enhancement:**
Add authentication layer:

**Option 1: JWT Authentication**
```protobuf
message UploadSnapshotRequest {
  string auth_token = 1;
  string filename = 2;
  bytes file_content = 3;
}
```

**Option 2: External Auth Proxy**
- Deploy behind nginx with auth_request
- Use OAuth2 proxy
- Integrate with existing SSO

**Option 3: API Keys**
- Generate per-user API keys
- Validate in gRPC interceptor
- Store in database with permissions

**References:**
- Documentation: README security recommendations

---

### 13. No Rate Limiting

**Severity:** Low
**Impact:** DoS Protection

**Description:**
There is no rate limiting on any endpoint. A single client can:
- Upload unlimited snapshots rapidly
- Spam GetHostHistory requests
- Trigger expensive diff operations repeatedly

**Potential Attack:**
```bash
# Flood the server
while true; do
  grpcurl -d '{"ip_address": "1.2.3.4"}' localhost:9090 hostdiff.HostService/GetHostHistory
done
```

**Recommended Enhancement:**
Implement rate limiting:

**Option 1: Middleware Rate Limiter**
```go
import "golang.org/x/time/rate"

func NewRateLimitInterceptor() grpc.UnaryServerInterceptor {
    limiter := rate.NewLimiter(10, 100) // 10 req/sec, burst 100
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {
        if !limiter.Allow() {
            return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
        }
        return handler(ctx, req)
    }
}
```

**Option 2: Nginx Rate Limiting**
```nginx
limit_req_zone $binary_remote_addr zone=grpc_limit:10m rate=10r/s;

location / {
    limit_req zone=grpc_limit burst=20;
    grpc_pass grpc://backend:8080;
}
```

**References:**
- File: `backend/cmd/server/main.go`
- Related: DoS protection, production hardening

---

### 14. No Soft Delete / Snapshot Deletion

**Severity:** Low
**Impact:** Feature Gap

**Description:**
Once a snapshot is uploaded, there is no way to delete it (except manual database modification).

**Use Cases for Deletion:**
- Remove accidentally uploaded snapshots
- Delete sensitive data
- Clean up test data
- Comply with data retention policies

**Recommended Enhancement:**
Add `DeleteSnapshot` RPC:

**Proto Definition:**
```protobuf
message DeleteSnapshotRequest {
  string snapshot_id = 1;
  bool force = 2;  // bypass soft delete
}

message DeleteSnapshotResponse {
  string message = 1;
}
```

**Implementation Options:**

**Option 1: Hard Delete**
```go
DELETE FROM snapshots WHERE id = ?
```

**Option 2: Soft Delete (Recommended)**
```sql
ALTER TABLE snapshots ADD COLUMN deleted_at TEXT;
CREATE INDEX idx_deleted_at ON snapshots(deleted_at);

-- Queries filter out soft-deleted
SELECT * FROM snapshots WHERE deleted_at IS NULL;
```

**References:**
- Proto: `proto/host_diff.proto`
- Related: Data management, compliance

---

### 15. No Bulk Upload Support

**Severity:** Low
**Impact:** User Experience

**Description:**
Users must upload snapshots one at a time. For large datasets (e.g., 100 snapshots), this is tedious.

**Current Limitation:**
- Frontend file input accepts one file
- No API for batch upload
- CLI requires scripting

**Recommended Enhancement:**

**Option 1: Batch Upload Endpoint**
```protobuf
message BatchUploadRequest {
  repeated UploadSnapshotRequest snapshots = 1;
}

message BatchUploadResponse {
  repeated UploadSnapshotResponse results = 1;
  int32 success_count = 2;
  int32 failure_count = 3;
}
```

**Option 2: Archive Upload**
Accept `.zip` or `.tar.gz` containing multiple snapshots:
```protobuf
message UploadArchiveRequest {
  bytes archive_content = 1;
  string format = 2;  // "zip" or "tar.gz"
}
```

**Option 3: Directory Upload (CLI)**
```bash
host-diff-cli upload --dir ./snapshots/
# Uploads all .json files in directory
```

**References:**
- File: `frontend/src/App.tsx:47-71`
- Proto: `proto/host_diff.proto`

---

## Planned Enhancements

### 16. Export Diff Reports (PDF, CSV, JSON)

**Priority:** Medium
**Effort:** Medium

**Description:**
Add ability to export diff comparison results in various formats for reporting and auditing.

**Proposed Formats:**

**1. PDF Report**
- Executive summary
- Visual charts (added/removed/changed)
- Detailed change tables
- Timestamp and metadata

**2. CSV Export**
```csv
Change Type,Port,Protocol,Field,Old Value,New Value
Added,443,HTTPS,,,
Modified,80,HTTP,status,200,301
Removed,22,SSH,,,
```

**3. JSON Export**
```json
{
  "comparison": {
    "snapshot_a": "1",
    "snapshot_b": "2",
    "timestamp": "2025-10-18T12:00:00Z"
  },
  "changes": [...]
}
```

**Implementation:**
- Add `ExportDiff` RPC
- Use libraries: `pdfcpu`, `encoding/csv`, `encoding/json`
- Frontend download button

**References:**
- Related: Reporting, compliance, auditing

---

### 17. Scheduled Comparisons and Alerting

**Priority:** Medium
**Effort:** High

**Description:**
Automatically compare latest snapshot against previous and send alerts on critical changes.

**Features:**
- Schedule: Compare every N hours/days
- Alert Rules: Define critical changes (e.g., new CVE, port closure)
- Notifications: Email, Slack, PagerDuty
- Alert History: Track all triggered alerts

**Example Configuration:**
```yaml
schedules:
  - name: "Daily Security Check"
    ip_addresses: ["192.168.1.1", "10.0.0.5"]
    frequency: "daily"
    alerts:
      - type: "new_cve"
        severity: "critical"
        channels: ["email", "slack"]
      - type: "port_removed"
        ports: [22, 443]
        severity: "high"
        channels: ["email"]
```

**Implementation:**
- Background worker process
- Alert rule engine
- Notification service integrations
- Admin API for managing schedules

**References:**
- Related: Automation, monitoring, DevOps integration

---

### 18. Semantic Version Comparison

**Priority:** Low
**Effort:** Low

**Description:**
Intelligently compare software versions using semantic versioning rules instead of string comparison.

**Current Behavior:**
```
"1.2.9" -> "1.2.10"  reported as change (correct)
"1.2.9" > "1.2.10"   (string comparison, incorrect)
```

**Enhanced Behavior:**
```
"1.2.9" -> "1.2.10"  = upgrade (minor)
"2.0.0" -> "1.9.9"   = downgrade (major)
"1.2.9" -> "1.2.9"   = no change
```

**Implementation:**
```go
import "github.com/hashicorp/go-version"

func compareVersions(oldVer, newVer string) string {
    v1, _ := version.NewVersion(oldVer)
    v2, _ := version.NewVersion(newVer)

    if v1.LessThan(v2) {
        return fmt.Sprintf("%s -> %s (upgrade)", oldVer, newVer)
    } else if v1.GreaterThan(v2) {
        return fmt.Sprintf("%s -> %s (downgrade)", oldVer, newVer)
    }
    return ""
}
```

**References:**
- File: `backend/internal/diff/diff.go:119-129`

---

### 19. Diff Visualization Improvements

**Priority:** Medium
**Effort:** Medium

**Description:**
Enhance the frontend diff viewer with better visualizations and filtering.

**Proposed Features:**

**1. Timeline View**
Show all snapshots on a timeline with visual indicators of change magnitude.

**2. Change Heatmap**
Visual representation of which services change most frequently.

**3. Filtering and Sorting**
- Filter by change type (added/removed/modified)
- Filter by port range
- Filter by severity (CVE-based)
- Sort by port, protocol, timestamp

**4. Side-by-Side Comparison**
```
Port 80 (HTTP)              Port 80 (HTTP)
  Status: 200         →       Status: 301
  Version: 1.2.3      →       Version: 1.2.4
  CVEs: []            →       CVEs: [CVE-2025-1234]
```

**5. Export View as Image**
Allow users to screenshot/export the diff view.

**References:**
- File: `frontend/src/DiffViewer.tsx`
- Related: UX improvements

---

### 20. Database Query Optimization

**Priority:** Low
**Effort:** Low

**Description:**
Add additional indexes and optimize queries for common access patterns.

**Current Indexes:**
```sql
CREATE INDEX idx_ip_timestamp ON snapshots(ip_address, timestamp DESC);
```

**Proposed Additional Indexes:**

**1. ID Lookups (for Compare)**
Already covered by PRIMARY KEY, but explicitly:
```sql
-- Already exists implicitly
-- CREATE INDEX idx_id ON snapshots(id);
```

**2. Timestamp-Only Queries**
```sql
CREATE INDEX idx_timestamp ON snapshots(timestamp DESC);
-- For queries like "get all snapshots in date range"
```

**3. Composite for Common Queries**
```sql
CREATE INDEX idx_ip_timestamp_id ON snapshots(ip_address, timestamp DESC, id);
-- Covering index for GetHostHistory
```

**4. Query Optimization**
Add `LIMIT` to queries returning multiple rows:
```go
func (db *DB) GetSnapshotsByIP(ipAddress string) ([]*Snapshot, error) {
    query := `
        SELECT id, ip_address, timestamp, data
        FROM snapshots
        WHERE ip_address = ? AND deleted_at IS NULL
        ORDER BY timestamp DESC
        LIMIT 1000  -- Prevent unbounded queries
    `
}
```

**References:**
- File: `backend/internal/data/database.go`
- Related: Performance, scalability

---

### 21. Multi-Tenant Support

**Priority:** Low
**Effort:** High

**Description:**
Add support for multiple organizations/teams with data isolation.

**Requirements:**
- User authentication (see #12)
- Organization/tenant ID
- Row-level security
- Admin UI for tenant management

**Database Schema Changes:**
```sql
CREATE TABLE organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL,
    role TEXT NOT NULL,  -- 'admin', 'user', 'viewer'
    FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

ALTER TABLE snapshots ADD COLUMN organization_id INTEGER NOT NULL;
CREATE INDEX idx_org_ip_timestamp ON snapshots(organization_id, ip_address, timestamp DESC);
```

**API Changes:**
```protobuf
message UploadSnapshotRequest {
  string auth_token = 1;       // JWT containing org_id
  string filename = 2;
  bytes file_content = 3;
}
```

All queries automatically filter by `organization_id` from auth token.

**References:**
- Related: Enterprise features, SaaS deployment

---

## Security Enhancements

### 22. Input Sanitization Hardening

**Priority:** Medium
**Effort:** Low

**Description:**
Add additional input sanitization beyond current validation.

**Enhancements:**

**1. Filename Path Traversal Prevention**
```go
func ParseFilename(filename string) (*ParsedFilename, error) {
    // Strip any directory paths
    filename = filepath.Base(filename)

    // Check for path traversal attempts
    if strings.Contains(filename, "..") {
        return nil, fmt.Errorf("invalid filename: path traversal detected")
    }

    // Existing validation...
}
```

**2. JSON Size Limits**
```go
const MaxSnapshotSize = 10 * 1024 * 1024 // 10MB

func (s *Server) UploadSnapshot(ctx context.Context, req *proto.UploadSnapshotRequest)
    (*proto.UploadSnapshotResponse, error) {
    if len(req.GetFileContent()) > MaxSnapshotSize {
        return nil, fmt.Errorf("snapshot too large: %d bytes (max %d)",
            len(req.GetFileContent()), MaxSnapshotSize)
    }
    // ...
}
```

**3. IP Address Allowlist/Blocklist**
```go
var blockedIPs = []string{
    "0.0.0.0",
    "255.255.255.255",
    "127.0.0.1",  // optional: block localhost
}

func validateIPAddress(ip string) error {
    for _, blocked := range blockedIPs {
        if ip == blocked {
            return fmt.Errorf("IP address %s is not allowed", ip)
        }
    }
    // ... existing validation
}
```

**References:**
- File: `backend/internal/validation/filename.go`
- File: `backend/internal/server/server.go`

---

### 23. Audit Logging

**Priority:** Medium
**Effort:** Medium

**Description:**
Log all operations for security auditing and compliance.

**Log Events:**
- Snapshot uploads (who, when, what IP)
- Snapshot comparisons (who compared what)
- Failed authentication attempts
- Database queries
- Configuration changes

**Implementation:**

**1. Structured Logging**
```go
import "go.uber.org/zap"

logger.Info("snapshot_uploaded",
    zap.String("user_id", userID),
    zap.String("ip_address", ipAddress),
    zap.String("snapshot_id", snapshotID),
    zap.Time("timestamp", time.Now()),
)
```

**2. Audit Table**
```sql
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    user_id TEXT,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT
);
```

**3. Export to SIEM**
- Support syslog output
- JSON log format for ingestion
- Integration with ELK stack

**References:**
- Related: Compliance, security, debugging

---

### 24. HTTPS/TLS for Production

**Priority:** High (for production)
**Effort:** Low

**Description:**
Enable TLS encryption for all communication.

**Current State:**
- gRPC uses plaintext (grpc.WithInsecure)
- Nginx serves HTTP only
- No certificate management

**Recommended Implementation:**

**1. Update Nginx Config**
```nginx
server {
    listen 443 ssl http2;
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # ... rest of config
}
```

**2. Update gRPC Server**
```go
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
if err != nil {
    log.Fatalf("Failed to load TLS keys: %v", err)
}
grpcServer := grpc.NewServer(grpc.Creds(creds))
```

**3. Certificate Management**
- Use Let's Encrypt for free certificates
- Implement auto-renewal
- Store certs in Docker secrets/volumes

**References:**
- File: `nginx.conf`
- File: `backend/cmd/server/main.go`

---

## Performance Optimizations

### 25. Caching Layer

**Priority:** Low
**Effort:** Medium

**Description:**
Add caching to reduce database load for frequently accessed data.

**Cache Targets:**

**1. Host History (High Cache Hit Rate)**
```go
// Cache GetHostHistory results for 5 minutes
cache.Set(fmt.Sprintf("history:%s", ipAddress), snapshots, 5*time.Minute)
```

**2. Diff Reports**
```go
// Cache comparison results (expensive operation)
cacheKey := fmt.Sprintf("diff:%s:%s", snapshotA, snapshotB)
cache.Set(cacheKey, report, 1*time.Hour)
```

**3. Snapshot Data**
```go
// Cache individual snapshot JSON
cache.Set(fmt.Sprintf("snapshot:%s", id), data, 15*time.Minute)
```

**Implementation Options:**

**Option 1: In-Memory Cache**
```go
import "github.com/patrickmn/go-cache"

var cache = cache.New(5*time.Minute, 10*time.Minute)
```

**Option 2: Redis**
```go
import "github.com/go-redis/redis/v8"

rdb := redis.NewClient(&redis.Options{
    Addr: "redis:6379",
})
```

**Cache Invalidation:**
- Clear cache on snapshot upload
- TTL-based expiration
- LRU eviction policy

**References:**
- Related: Performance, scalability

---

### 26. Database Connection Pooling Tuning

**Priority:** Low
**Effort:** Low

**Description:**
Current connection pool is set to 1 connection (SQLite limitation). For PostgreSQL migration, optimize pooling.

**Current Settings:**
```go
db.SetMaxOpenConns(1)
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(0)
```

**Recommended for PostgreSQL:**
```go
db.SetMaxOpenConns(25)        // Max concurrent connections
db.SetMaxIdleConns(5)         // Keep 5 connections alive
db.SetConnMaxLifetime(5*time.Minute)  // Recycle connections
db.SetConnMaxIdleTime(1*time.Minute)  // Close idle after 1min
```

**Monitoring:**
```go
stats := db.Stats()
log.Printf("DB Pool - Open: %d, InUse: %d, Idle: %d",
    stats.OpenConnections, stats.InUse, stats.Idle)
```

**References:**
- File: `backend/internal/data/database.go:19-47`

---

### 27. Incremental Diff Algorithm

**Priority:** Low
**Effort:** High

**Description:**
For very large snapshots (thousands of services), the current diff algorithm loads entire snapshots into memory and builds maps. This could be optimized.

**Current Approach:**
1. Load snapshot A (10MB) into memory
2. Load snapshot B (10MB) into memory
3. Build map A (all services)
4. Build map B (all services)
5. Compare

**Optimized Approach:**
1. Stream parse JSON (don't load entirely)
2. Use sorted services for merge-join algorithm
3. Process one service at a time

**Implementation:**
```go
func DiffSnapshotsIncremental(readerA, readerB io.Reader) (*DiffReport, error) {
    decoderA := json.NewDecoder(readerA)
    decoderB := json.NewDecoder(readerB)

    // Stream parse and compare
    // Similar to merge-join in databases
}
```

**Benefits:**
- Lower memory usage
- Faster for large snapshots
- Can handle snapshots > available RAM

**Tradeoffs:**
- More complex code
- Requires sorted input
- May be slower for small snapshots

**References:**
- File: `backend/internal/diff/diff.go:68-89`

---

## Developer Experience Improvements

### 28. Docker Compose Profiles

**Priority:** Low
**Effort:** Low

**Description:**
Add Docker Compose profiles for different deployment scenarios.

**Proposed Profiles:**

```yaml
services:
  backend:
    profiles: ["dev", "prod"]

  frontend:
    profiles: ["dev", "prod"]

  nginx:
    profiles: ["prod"]

  postgres:
    profiles: ["prod"]
    # PostgreSQL instead of SQLite for production

  redis:
    profiles: ["prod"]
    # Redis cache for production

  prometheus:
    profiles: ["monitoring"]

  grafana:
    profiles: ["monitoring"]
```

**Usage:**
```bash
# Development (minimal stack)
docker compose --profile dev up

# Production (full stack with monitoring)
docker compose --profile prod --profile monitoring up
```

**References:**
- File: `docker-compose.yml`

---

### 29. CLI Tool for Snapshot Management

**Priority:** Medium
**Effort:** Medium

**Description:**
Create a dedicated CLI tool for power users and automation.

**Proposed Features:**

```bash
# Upload snapshot
host-diff upload snapshots/host_192.168.1.1_2025-10-16T12-00-00Z.json

# Bulk upload
host-diff upload --dir ./snapshots/

# Get history
host-diff history 192.168.1.1

# Compare snapshots
host-diff compare --snapshot-a 123 --snapshot-b 456

# Export comparison
host-diff compare --snapshot-a 123 --snapshot-b 456 --format json > diff.json

# List all snapshots
host-diff list --ip 192.168.1.1 --limit 10

# Delete snapshot
host-diff delete --snapshot-id 123 --force
```

**Implementation:**
```go
package main

import (
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "host-diff"}

    uploadCmd := &cobra.Command{
        Use: "upload [file]",
        Run: uploadSnapshot,
    }

    rootCmd.AddCommand(uploadCmd)
    rootCmd.Execute()
}
```

**References:**
- Related: Developer tools, automation

---

### 30. OpenAPI/Swagger Documentation

**Priority:** Low
**Effort:** Medium

**Description:**
Generate OpenAPI documentation for the gRPC API using grpc-gateway.

**Implementation:**

**1. Add grpc-gateway Annotations**
```protobuf
import "google/api/annotations.proto";

service HostService {
  rpc UploadSnapshot(UploadSnapshotRequest) returns (UploadSnapshotResponse) {
    option (google.api.http) = {
      post: "/v1/snapshots"
      body: "*"
    };
  }
}
```

**2. Generate OpenAPI Spec**
```bash
protoc --openapiv2_out=. proto/host_diff.proto
```

**3. Serve Swagger UI**
```
http://localhost/swagger-ui
```

**Benefits:**
- Auto-generated API docs
- Interactive API testing
- REST endpoint generation (bonus)

**References:**
- Proto: `proto/host_diff.proto`
- Related: Documentation, REST support

---

## Summary Statistics

**Total Issues Documented:** 30

**By Severity:**
- Critical: 0
- High: 0
- Medium: 7
- Low: 23

**By Category:**
- Critical Issues: 3
- Known Bugs: 7
- Limitations: 5
- Planned Enhancements: 6
- Security Enhancements: 3
- Performance Optimizations: 3
- Developer Experience: 3

**Quick Wins (Low Effort, High Impact):**
1. #1 - Fix filename regex anchors (1 line change)
2. #7 - Add port number validation (5 lines)
3. #8 - Use time.Parse for timestamp validation (10 lines)
4. #22 - Add filepath.Base() for path traversal prevention (1 line)

**High-Priority Production Requirements:**
1. #12 - Authentication/Authorization
2. #13 - Rate Limiting
3. #24 - HTTPS/TLS
4. #23 - Audit Logging
5. #5 - Pagination for GetHostHistory

---

**Document Maintenance:**
This document should be updated whenever:
- New bugs are discovered
- Issues are fixed (move to changelog)
- Enhancements are implemented (move to changelog)
- Priorities change based on user feedback

**Contributing:**
When adding new issues, please include:
- Clear description of the problem
- Code references (file:line)
- Recommended fix or workaround
- Severity and impact assessment
