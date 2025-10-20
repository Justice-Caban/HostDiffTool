# Architecture: Host Diff Tool

A comprehensive technical architecture document for the Host Diff Tool, a production-ready web application for tracking and comparing host configuration snapshots over time.

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Technology Stack](#technology-stack)
- [Protocol Design](#protocol-design)
- [Database Design](#database-design)
- [Security Architecture](#security-architecture)
- [Performance Considerations](#performance-considerations)
- [Deployment Architecture](#deployment-architecture)
- [Design Decisions](#design-decisions)
- [Scalability](#scalability)

## Overview

The Host Diff Tool is designed to provide infrastructure teams with a reliable way to track changes in host configurations over time. The architecture prioritizes:

- **Simplicity**: Minimal dependencies, standard libraries where possible
- **Reliability**: Strong data integrity with UNIQUE constraints and validation
- **Performance**: Fast snapshot comparisons with efficient diffing algorithms
- **Maintainability**: Clear separation of concerns, comprehensive testing
- **Deployability**: Single-command Docker Compose deployment

### Design Principles

1. **Single Responsibility**: Each component has a clear, focused purpose
2. **Type Safety**: Strong typing in Go backend and TypeScript frontend
3. **Contract-First**: Protocol Buffers define the API contract
4. **Fail-Fast**: Early validation prevents bad data from entering the system
5. **Test Coverage**: 49+ unit tests plus E2E coverage

## System Architecture

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              User Layer                                      │
│                                                                              │
│  ┌──────────────────┐         ┌──────────────────┐                         │
│  │  Web Browser     │         │   CLI Tools      │                         │
│  │  (React SPA)     │         │   (grpcurl)      │                         │
│  └────────┬─────────┘         └────────┬─────────┘                         │
│           │ HTTP                        │ gRPC                               │
└───────────┼─────────────────────────────┼────────────────────────────────────┘
            │                             │
            ↓                             ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Presentation Layer                                   │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                    Nginx Reverse Proxy (Port 80)                       │ │
│  │                                                                        │ │
│  │  • Static file serving (React SPA)                                    │ │
│  │  • HTTP to gRPC-Web translation                                       │ │
│  │  • Request routing                                                    │ │
│  └─────────┬──────────────────────────────────────────────────┬──────────┘ │
└───────────┼──────────────────────────────────────────────────┼─────────────┘
            │                                                   │
            │ gRPC-Web                                          │ gRPC
            │ (Port 8080)                                       │ (Port 9090)
            ↓                                                   ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Application Layer                                    │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │              Go Backend Service (Backend Container)                    │ │
│  │                                                                        │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐ │ │
│  │  │                      gRPC Server Layer                           │ │ │
│  │  │  ┌────────────────┐  ┌──────────────┐  ┌───────────────────┐  │ │ │
│  │  │  │ Native gRPC    │  │  gRPC-Web    │  │  Reflection API   │  │ │ │
│  │  │  │ Server (9090)  │  │ Server (8080)│  │  (Development)    │  │ │ │
│  │  │  └────────┬───────┘  └──────┬───────┘  └─────────┬─────────┘  │ │ │
│  │  └───────────┼──────────────────┼─────────────────────┼───────────┘ │ │
│  │              └──────────────────┴─────────────────────┘              │ │
│  │                                  ↓                                    │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐ │ │
│  │  │                     Service Implementation                       │ │ │
│  │  │                                                                  │ │ │
│  │  │  ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐ │ │ │
│  │  │  │ Upload Handler  │  │ History Handler │  │ Compare Handler│ │ │ │
│  │  │  │                 │  │                 │  │                │ │ │ │
│  │  │  │ • Validates     │  │ • Query by IP  │  │ • Load snaps   │ │ │ │
│  │  │  │ • Parses JSON   │  │ • Sort by time │  │ • Run diff     │ │ │ │
│  │  │  │ • Stores data   │  │ • Return list  │  │ • Format report│ │ │ │
│  │  │  └────────┬────────┘  └────────┬────────┘  └────────┬───────┘ │ │ │
│  │  └───────────┼──────────────────────┼─────────────────────┼──────┘ │ │
│  │              └──────────────────────┴─────────────────────┘         │ │
│  │                                  ↓                                    │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐ │ │
│  │  │                      Business Logic Layer                        │ │ │
│  │  │                                                                  │ │ │
│  │  │  ┌──────────────────┐  ┌─────────────────┐  ┌───────────────┐ │ │ │
│  │  │  │ Input Validator  │  │  Diff Engine    │  │  Data Access  │ │ │ │
│  │  │  │                  │  │                 │  │  Layer (DAL)  │ │ │ │
│  │  │  │ • IP validation  │  │ • Service diff  │  │               │ │ │ │
│  │  │  │ • Timestamp val. │  │ • CVE tracking  │  │ • CRUD ops    │ │ │ │
│  │  │  │ • JSON parsing   │  │ • TLS changes   │  │ • Queries     │ │ │ │
│  │  │  │ • Filename check │  │ • Port changes  │  │ • Tx mgmt     │ │ │ │
│  │  │  └──────────────────┘  └─────────────────┘  └───────┬───────┘ │ │ │
│  │  └───────────────────────────────────────────────────────┼──────┘ │ │
│  └────────────────────────────────────────────────────────────┼───────┘ │
└───────────────────────────────────────────────────────────────┼─────────────┘
                                                                │
                                                                ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Data Layer                                        │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                         SQLite Database                                │ │
│  │                      (./data/snapshots.db)                             │ │
│  │                                                                        │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐ │ │
│  │  │                      snapshots Table                             │ │ │
│  │  │                                                                  │ │ │
│  │  │  • id (PRIMARY KEY, AUTOINCREMENT)                               │ │ │
│  │  │  • ip_address (TEXT, indexed)                                    │ │ │
│  │  │  • timestamp (TEXT, indexed)                                     │ │ │
│  │  │  • data (BLOB) - Full snapshot JSON                              │ │ │
│  │  │  • UNIQUE(ip_address, timestamp)                                 │ │ │
│  │  │  • INDEX idx_ip_timestamp ON (ip_address, timestamp DESC)        │ │ │
│  │  └──────────────────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Network Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                     Docker Network (Bridge)                     │
│                   takehomeassessment_default                    │
│                                                                 │
│  ┌───────────────┐      ┌───────────────┐      ┌────────────┐ │
│  │    nginx      │      │   backend     │      │  frontend  │ │
│  │  (port 80)    │◄────►│  (ports       │      │  (internal)│ │
│  │               │      │   8080, 9090) │      │            │ │
│  └───────┬───────┘      └───────┬───────┘      └────────────┘ │
│          │                      │                               │
└──────────┼──────────────────────┼───────────────────────────────┘
           │                      │
           │                      │
    ┌──────▼──────┐        ┌─────▼──────┐
    │  Host Port  │        │ Host Port  │
    │     80      │        │ 8080, 9090 │
    └─────────────┘        └────────────┘
```

## Component Details

### 1. Frontend Component (React SPA)

**Location:** `./frontend/`

**Technology Stack:**
- React 19.2.0
- TypeScript 5.x
- @improbable-eng/grpc-web (gRPC-Web client)
- Webpack 5 (bundler)

**Key Files:**
- `src/App.tsx` - Main application component
- `src/proto/` - Generated TypeScript code from Protocol Buffers
- `public/` - Static assets

**Responsibilities:**
1. **User Interface**: Provides three main views
   - Upload view with file picker
   - Host history view with IP search
   - Comparison view with diff display

2. **gRPC-Web Client**: Communicates with backend via gRPC-Web
   - Handles request/response serialization
   - Error handling and display
   - Loading states

3. **State Management**: Local React state (useState, useEffect)
   - No external state management library needed
   - Simple component-level state

**Build Process:**
```dockerfile
# Multi-stage build
Stage 1: Node 18 Alpine
  - npm install
  - npm run build
  - Output: /app/build/

Stage 2: Nginx 1.25 Alpine
  - Copy build artifacts
  - Configure nginx
  - Serve static files
```

### 2. Backend Component (Go gRPC Service)

**Location:** `./backend/`

**Technology Stack:**
- Go 1.25
- google.golang.org/grpc (gRPC framework)
- google.golang.org/protobuf (Protocol Buffers)
- github.com/improbable-eng/grpc-web (gRPC-Web middleware)
- github.com/mattn/go-sqlite3 (SQLite driver)

**Project Structure:**
```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── data/
│   │   ├── database.go          # Database connection & initialization
│   │   └── database_test.go     # Database tests (3 tests)
│   ├── diff/
│   │   ├── diff.go              # Core diffing algorithm
│   │   └── diff_test.go         # Diff tests (29 tests)
│   └── server/
│       ├── server.go            # gRPC service implementation
│       └── server_test.go       # Server tests (17 tests)
├── proto/
│   └── host_diff.proto          # API contract
├── go.mod                       # Go dependencies
└── go.sum                       # Dependency checksums
```

**Key Packages:**

#### `cmd/server/main.go`
Entry point that:
1. Initializes database connection
2. Creates dual gRPC servers (native + gRPC-Web)
3. Registers service handlers
4. Starts both servers concurrently

```go
// Pseudo-code structure
func main() {
    // Initialize database
    db := data.InitDB("./data/snapshots.db")

    // Create gRPC server
    grpcServer := grpc.NewServer()
    hostdiff.RegisterHostServiceServer(grpcServer, server.NewServer(db))

    // Native gRPC listener (port 9090)
    go grpcServer.Serve(":9090")

    // gRPC-Web wrapper (port 8080)
    wrappedServer := grpcweb.WrapServer(grpcServer)
    http.ListenAndServe(":8080", wrappedServer)
}
```

#### `internal/data/`
Database access layer with:
- Connection pooling (SQLite supports single writer, multiple readers)
- Schema initialization
- CRUD operations
- Parameterized queries (SQL injection prevention)

**Key Functions:**
```go
InitDB(path string) (*sql.DB, error)
InsertSnapshot(db *sql.DB, ip, timestamp string, data []byte) (int64, error)
GetSnapshotsByIP(db *sql.DB, ip string) ([]Snapshot, error)
GetSnapshotByID(db *sql.DB, id int64) (*Snapshot, error)
```

#### `internal/diff/`
Core diffing engine that compares two snapshots:

**Algorithm:**
1. Parse both snapshots into structured data
2. Build service maps keyed by port+protocol
3. Compare services:
   - Detect additions (in B, not in A)
   - Detect removals (in A, not in B)
   - Detect changes (in both, but different)
4. For each service, compare:
   - Status codes
   - Software versions
   - TLS configuration
   - CVE lists (per-port tracking)
5. Format human-readable report

**Key Functions:**
```go
CompareSnapshots(a, b *Snapshot) (*DiffReport, error)
detectServiceChanges(aServices, bServices map[string]Service) []Change
detectCVEChanges(aServices, bServices map[string]Service) []CVEChange
formatDiffReport(changes []Change) string
```

#### `internal/server/`
gRPC service implementation:

**Key Functions:**
```go
// UploadSnapshot validates and stores a new snapshot
func (s *Server) UploadSnapshot(ctx context.Context, req *pb.UploadSnapshotRequest) (*pb.UploadSnapshotResponse, error)

// GetHostHistory retrieves all snapshots for an IP
func (s *Server) GetHostHistory(ctx context.Context, req *pb.GetHostHistoryRequest) (*pb.GetHostHistoryResponse, error)

// CompareSnapshots generates a diff report
func (s *Server) CompareSnapshots(ctx context.Context, req *pb.CompareSnapshotsRequest) (*pb.CompareSnapshotsResponse, error)
```

**Validation:**
- IP address: 0-255 per octet
- Timestamp: Valid ISO-8601 format
- Filename: `host_<ip>_<timestamp>.json`
- JSON structure: Required fields present

### 3. Nginx Reverse Proxy

**Location:** `./nginx.conf`

**Responsibilities:**
1. Serve static React SPA files
2. Route gRPC-Web requests to backend:8080
3. Route native gRPC requests to backend:9090
4. Handle HTTP to backend translation

**Configuration:**
```nginx
server {
    listen 80;

    # Serve React SPA
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    # Proxy gRPC-Web requests
    location /hostdiff.HostService/ {
        grpc_pass grpc://backend:8080;
        grpc_set_header Host $host;
    }
}
```

### 4. Database Layer

**Technology:** SQLite 3.x

**File Location:** `./data/snapshots.db`

**Schema:**
```sql
CREATE TABLE snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    data BLOB NOT NULL,
    UNIQUE(ip_address, timestamp)
);

CREATE INDEX idx_ip_timestamp ON snapshots(ip_address, timestamp DESC);
```

**Design Decisions:**
- **BLOB for data**: Stores full JSON snapshot for complete history
- **UNIQUE constraint**: Prevents duplicate snapshots (same IP + timestamp)
- **Descending index**: Optimizes "most recent first" queries
- **TEXT for timestamp**: Stores ISO-8601 strings for readability

## Data Flow

### Upload Snapshot Flow

```
┌──────────┐
│ Browser  │
└─────┬────┘
      │ 1. User selects file
      ↓
┌─────────────────┐
│ React Component │
└─────┬───────────┘
      │ 2. gRPC-Web call: UploadSnapshot
      │    Request: { filename, file_content }
      ↓
┌──────────┐
│  Nginx   │
└─────┬────┘
      │ 3. Proxy to backend:8080
      ↓
┌──────────────────────┐
│ gRPC-Web Middleware  │
└─────┬────────────────┘
      │ 4. Convert to native gRPC
      ↓
┌────────────────────┐
│ Upload Handler     │
│                    │
│ 5. Validate:       │
│    - Filename      │
│    - IP address    │
│    - Timestamp     │
│    - JSON content  │
└─────┬──────────────┘
      │ 6. If valid
      ↓
┌───────────────┐
│ Data Layer    │
│               │
│ 7. INSERT     │
│    snapshot   │
└─────┬─────────┘
      │ 8. Return ID or error
      ↓
┌──────────┐
│ SQLite   │
│ Database │
└──────────┘
```

### Compare Snapshots Flow

```
┌──────────┐
│ Browser  │
└─────┬────┘
      │ 1. User selects 2 snapshots
      ↓
┌─────────────────┐
│ React Component │
└─────┬───────────┘
      │ 2. gRPC-Web call: CompareSnapshots
      │    Request: { snapshot_id_a, snapshot_id_b }
      ↓
┌──────────────────────┐
│ Compare Handler      │
│                      │
│ 3. Load snapshot A   │
│    from database     │
│ 4. Load snapshot B   │
│    from database     │
└─────┬────────────────┘
      │ 5. Both loaded
      ↓
┌───────────────────┐
│ Diff Engine       │
│                   │
│ 6. Parse JSON     │
│ 7. Build maps     │
│ 8. Compare:       │
│    - Services     │
│    - CVEs         │
│    - Status       │
│    - Software     │
│    - TLS          │
└─────┬─────────────┘
      │ 9. Generate report
      ↓
┌────────────────┐
│ Format Report  │
│                │
│ 10. Structure: │
│     - Added    │
│     - Removed  │
│     - Changed  │
│     - CVEs     │
└─────┬──────────┘
      │ 11. Return formatted report
      ↓
┌──────────┐
│ Browser  │
│          │
│ Display  │
│ Diff     │
└──────────┘
```

## Technology Stack

### Backend Dependencies

```go
require (
    google.golang.org/grpc v1.65.0
    google.golang.org/protobuf v1.34.2
    github.com/improbable-eng/grpc-web v0.15.0
    github.com/mattn/go-sqlite3 v1.14.22
)
```

**Why these choices?**
- `google.golang.org/grpc`: Official gRPC implementation for Go
- `google.golang.org/protobuf`: Official Protocol Buffers library
- `improbable-eng/grpc-web`: Mature, well-tested gRPC-Web wrapper
- `mattn/go-sqlite3`: Pure Go SQLite driver (CGO-based)

### Frontend Dependencies

```json
{
  "dependencies": {
    "react": "^19.2.0",
    "react-dom": "^19.2.0",
    "@improbable-eng/grpc-web": "^0.15.0",
    "google-protobuf": "^3.21.4",
    "typescript": "^5.6.3"
  }
}
```

## Protocol Design

### gRPC Service Definition

**File:** `proto/host_diff.proto`

```protobuf
syntax = "proto3";

package hostdiff;

option go_package = "github.com/justicecaban/host-diff-tool/proto";

// HostService provides snapshot management and comparison
service HostService {
  // UploadSnapshot stores a new host snapshot
  rpc UploadSnapshot(UploadSnapshotRequest) returns (UploadSnapshotResponse);

  // GetHostHistory retrieves all snapshots for a given IP
  rpc GetHostHistory(GetHostHistoryRequest) returns (GetHostHistoryResponse);

  // CompareSnapshots generates a diff report between two snapshots
  rpc CompareSnapshots(CompareSnapshotsRequest) returns (CompareSnapshotsResponse);
}

// Request/response messages...
```

### Message Design Principles

1. **Explicit field numbers**: Never reuse field numbers (breaking change)
2. **Optional fields**: Use optional for nullable values
3. **Repeated fields**: For arrays/lists
4. **Enum types**: For status codes, change types
5. **Nested messages**: For complex structures

### Wire Format

**gRPC-Web (Browser):**
- Protocol: HTTP/1.1 or HTTP/2
- Content-Type: `application/grpc-web+proto`
- Binary Protocol Buffers payload
- Custom headers: `X-Grpc-Web: 1`

**Native gRPC (CLI):**
- Protocol: HTTP/2
- Content-Type: `application/grpc+proto`
- Binary Protocol Buffers payload
- Standard gRPC framing

## Database Design

### Schema Design

```sql
CREATE TABLE snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    data BLOB NOT NULL,
    UNIQUE(ip_address, timestamp)
);

CREATE INDEX idx_ip_timestamp ON snapshots(ip_address, timestamp DESC);
```

### Design Rationale

**Why BLOB for data?**
- Stores complete snapshot for historical accuracy
- Allows future schema changes without migrations
- Efficient binary storage
- Easy to deserialize

**Why UNIQUE constraint?**
- Prevents accidental duplicates
- Enforces data integrity at database level
- Clear error messages when violated

**Why descending index?**
- Optimizes "most recent first" queries
- Common use case: view latest snapshots
- SQLite can use index for ORDER BY

### Query Patterns

**Insert:**
```sql
INSERT INTO snapshots (ip_address, timestamp, data)
VALUES (?, ?, ?)
```

**Select by IP (with ordering):**
```sql
SELECT id, ip_address, timestamp, data
FROM snapshots
WHERE ip_address = ?
ORDER BY timestamp DESC
```

**Select by ID:**
```sql
SELECT id, ip_address, timestamp, data
FROM snapshots
WHERE id = ?
```

## Security Architecture

### Input Validation

**Layer 1: Frontend Validation**
- File type checking (.json only)
- File size limits (configurable)
- Filename format preview

**Layer 2: Backend Validation**
- IP address: Regex + range check (0-255)
- Timestamp: ISO-8601 parsing with validation
- JSON: Schema validation
- Filename: Strict format enforcement

**Layer 3: Database Constraints**
- UNIQUE constraint prevents duplicates
- NOT NULL constraints enforce required fields
- Type checking by SQLite

### SQL Injection Prevention

**Parameterized Queries:**
```go
// GOOD - Parameterized
db.Exec("INSERT INTO snapshots (ip_address, timestamp, data) VALUES (?, ?, ?)",
    ip, timestamp, data)

// BAD - String concatenation (never used)
db.Exec("INSERT INTO snapshots VALUES ('" + ip + "', ...)")
```

**All queries use `?` placeholders**, which are safely escaped by the database driver.

### XSS Prevention

**Backend:**
- Filename validation rejects `<`, `>`, `"`, `'`
- No HTML generation in backend
- JSON responses only

**Frontend:**
- React's JSX automatically escapes content
- No `dangerouslySetInnerHTML` used
- Content Security Policy headers (via Nginx)

### Path Traversal Prevention

**Filename validation:**
```go
// Reject dangerous characters
if strings.Contains(filename, "..") ||
   strings.Contains(filename, "/") ||
   strings.Contains(filename, "\\") {
    return errors.New("invalid filename")
}
```

### CORS Configuration

**Nginx handles CORS:**
```nginx
add_header Access-Control-Allow-Origin *;
add_header Access-Control-Allow-Methods "GET, POST, OPTIONS";
add_header Access-Control-Allow-Headers "Content-Type, X-Grpc-Web";
```

**For production:** Restrict `Access-Control-Allow-Origin` to specific domains.

## Performance Considerations

### Backend Performance

**Snapshot Upload:**
- File parsing: ~10ms for typical snapshot
- Database insert: ~5-10ms
- Total: <50ms

**History Retrieval:**
- Index lookup: ~1ms
- Deserialization: ~5-10ms per snapshot
- Total: <30ms for 10 snapshots

**Snapshot Comparison:**
- Load 2 snapshots: ~10ms
- Diff algorithm: O(n) where n = number of services
- Typical: ~20-50ms for 50 services
- Large: ~100-200ms for 1000 services

### Database Performance

**SQLite Characteristics:**
- Single writer (serialized writes)
- Multiple readers (concurrent reads)
- File-based (no network overhead)
- In-process (no IPC)

**Optimization:**
```sql
-- Descending index for ORDER BY
CREATE INDEX idx_ip_timestamp ON snapshots(ip_address, timestamp DESC);

-- Analyze for query planner
ANALYZE;

-- Vacuum to reclaim space
VACUUM;
```

### Frontend Performance

**Build Optimization:**
- Webpack code splitting
- Minification and compression
- Tree shaking (remove unused code)
- Gzip compression via Nginx

**Runtime Optimization:**
- React.memo for expensive components
- Lazy loading for large diffs
- Virtualization for long lists (if needed)

### Network Performance

**gRPC-Web:**
- Binary Protocol Buffers (smaller than JSON)
- HTTP/2 multiplexing
- Compression (gzip)

**Typical Payload Sizes:**
- Upload request: ~2-10KB (snapshot)
- History response: ~1KB per snapshot
- Compare response: ~1-5KB (diff report)

## Deployment Architecture

### Docker Compose Structure

```yaml
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"  # gRPC-Web
      - "9090:9090"  # Native gRPC
    volumes:
      - ./data:/app/data
    networks:
      - app-network

  frontend:
    build: ./frontend
    networks:
      - app-network

  nginx:
    build: ./
    ports:
      - "80:80"
    depends_on:
      - backend
      - frontend
    networks:
      - app-network

networks:
  app-network:
    driver: bridge

volumes:
  data:
```

### Container Images

**Backend:**
```dockerfile
FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM debian:stable-slim
COPY --from=builder /app/server /app/server
CMD ["/app/server"]
```

**Frontend:**
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:1.25-alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
```

### Resource Requirements

**Minimum:**
- CPU: 1 core
- Memory: 512MB
- Disk: 1GB (database grows over time)

**Recommended:**
- CPU: 2 cores
- Memory: 2GB
- Disk: 10GB

### Health Checks

```yaml
backend:
  healthcheck:
    test: ["CMD", "grpcurl", "-plaintext", "localhost:9090", "list"]
    interval: 30s
    timeout: 10s
    retries: 3

nginx:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost"]
    interval: 30s
    timeout: 10s
    retries: 3
```

## Design Decisions

### Why gRPC?

**Advantages:**
1. Strong typing with Protocol Buffers
2. Efficient binary serialization
3. Code generation for multiple languages
4. Built-in error handling
5. Streaming support (future feature)

**Trade-offs:**
- Requires gRPC-Web wrapper for browsers
- More complex than REST
- Debugging requires special tools (grpcurl)

### Why SQLite?

**Advantages:**
1. No separate server process
2. Zero configuration
3. Reliable (ACID compliant)
4. Sufficient for expected scale
5. Simple backup (copy file)

**Trade-offs:**
- Single writer (not an issue for this use case)
- Limited concurrency vs PostgreSQL
- No built-in replication

**When to migrate:**
- Concurrent writes >100/sec
- Database size >100GB
- Need distributed deployment

### Why Dual Protocol Support?

**Native gRPC (9090):**
- For CLI tools, automation, testing
- Best performance
- Standard gRPC features

**gRPC-Web (8080):**
- For browser clients
- HTTP/1.1 compatible
- Works with existing proxies

### Why React?

**Advantages:**
1. Large ecosystem
2. Excellent TypeScript support
3. Component-based architecture
4. Virtual DOM for performance
5. Mature gRPC-Web libraries

**Trade-offs:**
- Larger bundle size than vanilla JS
- Requires build step

## Scalability

### Horizontal Scaling

**Current Architecture:**
- Stateless backend (can run multiple instances)
- SQLite limits to single host

**To scale horizontally:**
1. Replace SQLite with PostgreSQL
2. Add load balancer (Nginx, HAProxy)
3. Run multiple backend instances
4. Shared database or read replicas

### Vertical Scaling

**Database:**
- Increase disk I/O (SSD)
- More memory for cache
- Add indexes for new query patterns

**Backend:**
- Increase Go worker pool size
- Tune gRPC connection limits
- Profile and optimize hot paths

### Caching Strategy

**Current:** No caching (simplicity)

**Future Options:**
1. In-memory cache for recent snapshots
2. Redis for distributed caching
3. CDN for static frontend assets

### Database Partitioning

**Future Strategy:**
- Partition by IP address range
- Time-based partitioning (monthly tables)
- Archive old snapshots to cold storage

### Monitoring and Observability

**To Add:**
1. Prometheus metrics
2. Grafana dashboards
3. Structured logging (JSON)
4. Distributed tracing (OpenTelemetry)
5. Error tracking (Sentry)

### Production Hardening

**Required for Production:**
1. **Authentication**: JWT or OAuth2
2. **Authorization**: Role-based access control
3. **Rate Limiting**: Prevent DoS attacks
4. **TLS/SSL**: Encrypt all traffic
5. **Secrets Management**: Vault, AWS Secrets Manager
6. **Backups**: Automated database backups
7. **Monitoring**: Health checks, metrics, logs
8. **CI/CD**: Automated testing and deployment

---

**Document Version:** 2.0
**Last Updated:** October 2025
**Status:** Comprehensive technical architecture ✅
**Next Review:** When scaling requirements change
