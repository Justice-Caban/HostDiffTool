# Troubleshooting Guide

This guide covers common issues and their solutions for the Host Diff Tool.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Startup Issues](#startup-issues)
- [Connection Problems](#connection-problems)
- [Database Issues](#database-issues)
- [Upload Problems](#upload-problems)
- [Performance Issues](#performance-issues)
- [Browser Issues](#browser-issues)
- [Testing Problems](#testing-problems)
- [Getting Help](#getting-help)

## Quick Diagnostics

Before diving into specific issues, run these quick checks:

```bash
# 1. Check Docker is running
docker ps

# 2. Check services status
docker compose ps

# 3. Check container logs
docker compose logs --tail=50

# 4. Check port availability
sudo lsof -i :80
sudo lsof -i :8080
sudo lsof -i :9090

# 5. Test web UI
curl -I http://localhost

# 6. Test gRPC endpoint
grpcurl -plaintext localhost:9090 list
```

**All services should show status "Up" and respond to requests.**

## Startup Issues

### Problem: "docker compose" command not found

**Symptoms:**
```bash
$ docker compose up
zsh: command not found: docker compose
```

**Solution 1: Try with hyphen (older version)**
```bash
docker-compose up --build
```

**Solution 2: Install Docker Compose v2**
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker-compose-plugin

# macOS
brew install docker-compose

# Verify
docker compose version
```

### Problem: Port 80 already in use

**Symptoms:**
```
Error starting userland proxy: listen tcp4 0.0.0.0:80: bind: address already in use
```

**Check what's using the port:**
```bash
sudo lsof -i :80
# Or
sudo netstat -tulpn | grep :80
```

**Solution 1: Stop the conflicting service**
```bash
# Apache
sudo systemctl stop apache2
sudo systemctl disable apache2

# Nginx
sudo systemctl stop nginx
sudo systemctl disable nginx

# Generic process
sudo kill <PID>
```

**Solution 2: Change the port**

Edit `docker-compose.yml`:
```yaml
services:
  nginx:
    ports:
      - "8000:80"  # Change 80 to 8000
```

Then access via `http://localhost:8000`

### Problem: Permission denied errors

**Symptoms:**
```
permission denied while trying to connect to the Docker daemon socket
```

**Solution 1: Add user to docker group**
```bash
sudo usermod -aG docker $USER
newgrp docker

# Or logout and login again
```

**Solution 2: Use sudo**
```bash
sudo docker compose up --build
```

### Problem: Services keep restarting

**Check logs:**
```bash
docker compose logs backend
docker compose logs frontend
docker compose logs nginx
```

**Common causes:**

**1. Backend can't create database:**
```bash
# Check permissions on data directory
ls -la data/

# Fix permissions
sudo chown -R $USER:$USER data/
```

**2. Frontend build failed:**
```bash
# Rebuild frontend
docker compose build --no-cache frontend
docker compose up frontend
```

**3. Port conflicts:**
```bash
# Check for conflicts
docker compose down
sudo lsof -i :8080
sudo lsof -i :9090
```

### Problem: "no space left on device"

**Check Docker disk usage:**
```bash
docker system df
```

**Clean up:**
```bash
# Remove stopped containers
docker container prune

# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Nuclear option (removes everything)
docker system prune -a --volumes
```

## Connection Problems

### Problem: Can't access web UI at http://localhost

**1. Check if nginx is running:**
```bash
docker compose ps nginx
```

**2. Check nginx logs:**
```bash
docker compose logs nginx
```

**3. Verify port mapping:**
```bash
docker compose port nginx 80
# Should output: 0.0.0.0:80
```

**4. Test with curl:**
```bash
curl -v http://localhost
```

**5. Try specific IP:**
```bash
curl -v http://127.0.0.1
curl -v http://$(hostname -I | awk '{print $1}')
```

**6. Check firewall:**
```bash
# Ubuntu/Debian
sudo ufw status
sudo ufw allow 80/tcp

# Firewalld
sudo firewall-cmd --list-all
sudo firewall-cmd --add-port=80/tcp --permanent
sudo firewall-cmd --reload
```

### Problem: gRPC connection refused

**Symptoms:**
```
Failed to dial target host "localhost:9090": context deadline exceeded
```

**1. Check backend is running:**
```bash
docker compose ps backend
```

**2. Check backend logs:**
```bash
docker compose logs backend
```

**Expected output:**
```
Starting native gRPC server on :9090
Starting gRPC-Web HTTP server on :8080
```

**3. Verify port exposure:**
```bash
docker compose port backend 9090
# Should output: 0.0.0.0:9090
```

**4. Test with grpcurl:**
```bash
# List services
grpcurl -plaintext localhost:9090 list

# Should output:
# grpc.reflection.v1alpha.ServerReflection
# hostdiff.HostService
```

**5. Check if port is actually listening:**
```bash
sudo netstat -tulpn | grep 9090
# Or
sudo lsof -i :9090
```

### Problem: gRPC-Web not working from browser

**1. Check browser console for errors (F12)**

**2. Verify backend gRPC-Web server:**
```bash
docker compose logs backend | grep "8080"
# Should show: Starting gRPC-Web HTTP server on :8080
```

**3. Test gRPC-Web endpoint:**
```bash
curl -X POST http://localhost:8080/hostdiff.HostService/GetHostHistory \
  -H "Content-Type: application/grpc-web+proto" \
  -H "X-Grpc-Web: 1"
```

**4. Check nginx configuration:**
```bash
docker compose exec nginx cat /etc/nginx/conf.d/default.conf
```

**5. Verify frontend can reach backend:**
```bash
docker compose exec frontend ping backend
```

## Database Issues

### Problem: "UNIQUE constraint failed"

**Symptoms:**
```
Error: failed to insert snapshot: UNIQUE constraint failed: snapshots.ip_address, snapshots.timestamp
```

**This means you're trying to upload a duplicate snapshot.**

**Solution 1: This is correct behavior (by design)**

**Solution 2: Clean database for testing:**
```bash
# Stop services
docker compose down -v

# Remove database
rm -rf data

# Restart
docker compose up -d
```

### Problem: "database is locked"

**Symptoms:**
```
Error: database is locked
```

**Solution:**
```bash
# Stop all services
docker compose down

# Remove database lock
rm -f data/snapshots.db-wal
rm -f data/snapshots.db-shm

# Restart
docker compose up -d
```

### Problem: Database corruption

**Symptoms:**
```
Error: database disk image is malformed
```

**Solution:**
```bash
# 1. Stop services
docker compose down

# 2. Backup existing database
cp data/snapshots.db data/snapshots.db.backup

# 3. Try to repair
sqlite3 data/snapshots.db "PRAGMA integrity_check;"

# 4. If repair fails, start fresh
rm -f data/snapshots.db
rm -f data/snapshots.db-wal
rm -f data/snapshots.db-shm

# 5. Restart
docker compose up -d
```

**Note:** The database now uses WAL (Write-Ahead Logging) mode for better performance and concurrency. This creates additional files (-wal and -shm) which are normal.

### Problem: Can't find database file

**Database location:**
```
./data/snapshots.db
```

**Check if it exists:**
```bash
ls -la data/
```

**If missing:**
```bash
# It will be created automatically on first run
docker compose restart backend
sleep 3

# Verify creation
ls -la data/
```

**Check permissions:**
```bash
ls -la data/
# data directory should be writable

# Fix if needed
sudo chown -R $USER:$USER data/
chmod 755 data/
```

## Upload Problems

### Problem: "invalid filename" error

**Valid filename format:**
```
host_<ip>_<timestamp>.json
```

**Examples:**
- ✅ `host_192.0.2.1_2025-10-17T12-00-00Z.json`
- ❌ `snapshot_192.0.2.1.json`
- ❌ `host_192.0.2.1.json`
- ❌ `192.0.2.1_2025-10-17.json`

**Validation rules:**
- **IP address**: Valid IPv4 (0-255 per octet)
- **Timestamp**: ISO-8601 format, use dashes not colons in the time portion
- **Extension**: Must be `.json`

**Common errors:**

1. **Invalid IP octet:**
   ```
   Error: invalid IP address octet [0]: 256 (must be 0-255)
   ```
   Fix: Ensure all IP octets are between 0-255

2. **Invalid timestamp:**
   ```
   Error: invalid month: 13 (must be 1-12)
   ```
   Fix: Use valid month (1-12), day (1-31), hour (0-23), minute/second (0-59)

3. **Wrong filename format:**
   ```
   Error: filename does not match expected format 'host_<ip>_<timestamp>.json'
   ```
   Fix: Must start with `host_`, include valid IP and timestamp

### Problem: "invalid JSON content" error

**Check JSON validity:**
```bash
# Test with jq
cat yourfile.json | jq .

# Or with Python
python3 -m json.tool yourfile.json
```

**Required JSON structure:**
```json
{
  "ip": "192.0.2.1",
  "timestamp": "2025-10-17T12:00:00Z",
  "services": [
    {
      "port": 80,
      "protocol": "HTTP"
    }
  ]
}
```

### Problem: Upload succeeds but nothing happens

**Check browser console (F12) for errors**

**Verify upload actually succeeded:**
```bash
# Check backend logs
docker compose logs backend | grep -i upload

# Or query database
docker exec -it takehomeassessment-backend-1 sh
sqlite3 /app/data/snapshots.db "SELECT COUNT(*) FROM snapshots;"
```

### Problem: File upload button doesn't work

**Browser console checks:**
1. Open DevTools (F12)
2. Check Console tab for JavaScript errors
3. Check Network tab to see if requests are being made

**Try different browser:**
- Chrome/Chromium
- Firefox
- Edge

**Clear browser cache:**
```
Ctrl+Shift+Delete (or Cmd+Shift+Delete on Mac)
```

**Rebuild frontend:**
```bash
docker compose build --no-cache frontend
docker compose up -d frontend nginx
```

## Performance Issues

### Problem: Slow uploads

**Check file size:**
```bash
ls -lh yourfile.json
```

**Large files (>1MB) may take longer**

**Check backend CPU/memory:**
```bash
docker stats takehomeassessment-backend-1
```

**Check database size:**
```bash
ls -lh data/snapshots.db
```

**Optimize database:**
```bash
docker exec -it takehomeassessment-backend-1 sh
sqlite3 /app/data/snapshots.db "VACUUM;"
sqlite3 /app/data/snapshots.db "ANALYZE;"
```

### Problem: Slow comparisons

**Expected times:**
- Small snapshots (<10 services): <50ms
- Medium snapshots (10-100 services): <200ms
- Large snapshots (100-1000 services): <1000ms

**If slower:**

**1. Check backend resources:**
```bash
docker stats
```

**2. Increase Docker resources:**

Edit Docker Desktop settings:
- CPUs: 4+
- Memory: 4GB+

**3. Check database size:**
```bash
# If >100MB, consider cleanup
ls -lh data/snapshots.db
```

### Problem: Web UI sluggish

**1. Check browser resources (Task Manager)**

**2. Clear browser cache**

**3. Check nginx logs:**
```bash
docker compose logs nginx
```

**4. Restart services:**
```bash
docker compose restart
```

## Browser Issues

### Problem: Page won't load

**1. Hard refresh:**
```
Ctrl+Shift+R (or Cmd+Shift+R on Mac)
```

**2. Clear browser cache:**
```
Ctrl+Shift+Delete
```

**3. Try incognito/private mode**

**4. Check browser console (F12) for errors**

### Problem: "net::ERR_CONNECTION_REFUSED"

**Backend is not accessible from frontend**

**1. Check all services running:**
```bash
docker compose ps
```

**2. Check Docker network:**
```bash
docker network ls
docker network inspect takehomeassessment_default
```

**3. Restart everything:**
```bash
docker compose down
docker compose up -d
```

### Problem: CORS errors

**Symptoms in browser console:**
```
Access to XMLHttpRequest blocked by CORS policy
```

**This shouldn't happen with our setup, but if it does:**

**Check nginx configuration:**
```bash
docker compose exec nginx cat /etc/nginx/conf.d/default.conf
```

**Verify proxy settings are correct:**
```nginx
location /hostdiff.HostService/ {
    grpc_pass backend:8080;
}
```

**Restart nginx:**
```bash
docker compose restart nginx
```

## Testing Problems

### Problem: e2e_test.sh fails with "UNIQUE constraint"

**Clean database before testing:**
```bash
rm -f data/snapshots.db
docker compose restart backend
sleep 5
./e2e_test.sh
```

### Problem: grpcurl command not found

**Install grpcurl:**
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Add to PATH:**
```bash
export PATH=$PATH:~/go/bin

# Make permanent
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Problem: Browser test fails to start

**Install Puppeteer dependencies:**
```bash
# Ubuntu/Debian
sudo apt-get install -y \
  chromium-browser \
  libnss3 \
  libatk-bridge2.0-0 \
  libgtk-3-0

# macOS (no action needed)

# Check if installed
npm list puppeteer
```

### Problem: Unit tests fail

**Clean and rebuild:**
```bash
cd backend
go clean -cache
go mod tidy
go mod download
go test ./...
```

## Getting Help

### Gathering Debug Information

When asking for help, provide:

```bash
# 1. System information
uname -a
docker --version
docker compose version

# 2. Service status
docker compose ps

# 3. Recent logs
docker compose logs --tail=100 > logs.txt

# 4. Container inspect
docker inspect takehomeassessment-backend-1 > backend-inspect.txt

# 5. Database state
docker exec takehomeassessment-backend-1 \
  sqlite3 /app/data/snapshots.db \
  "SELECT COUNT(*) FROM snapshots;"
```

### Common Log Patterns

**Success patterns:**
```
✓ Starting native gRPC server on :9090
✓ Starting gRPC-Web HTTP server on :8080
✓ Database initialized successfully
```

**Error patterns:**
```
❌ failed to listen on port
❌ database is locked
❌ UNIQUE constraint failed
❌ connection refused
```

### Reset Everything

**Nuclear option - start completely fresh:**

```bash
# 1. Stop everything
docker compose down -v

# 2. Remove all data
rm -rf data
rm -f e2e_test_screenshot.png

# 3. Clean Docker
docker system prune -f

# 4. Rebuild from scratch
docker compose build --no-cache

# 5. Start fresh
docker compose up -d

# 6. Wait for startup
sleep 10

# 7. Verify
docker compose ps
curl http://localhost
```

## Performance Troubleshooting

### Problem: Database is slow after many operations

**Symptoms:**
- Queries taking longer than usual
- Large database file size

**Solution - Run database maintenance:**
```bash
docker exec -it takehomeassessment-backend-1 sh
sqlite3 /app/data/snapshots.db

-- Check database size and fragmentation
.dbinfo

-- Optimize database (reclaim space)
VACUUM;

-- Update query planner statistics
ANALYZE;

-- Check that WAL mode is enabled
PRAGMA journal_mode;
-- Should return: wal

-- Check cache size
PRAGMA cache_size;
-- Should return: -64000 (64MB)

.exit
exit
```

**Check database files:**
```bash
ls -lh data/
# You should see:
# snapshots.db      (main database)
# snapshots.db-wal  (write-ahead log - normal with WAL mode)
# snapshots.db-shm  (shared memory - normal with WAL mode)
```

### Problem: Frontend loads slowly

**Check if it's a caching issue:**
```bash
# Clear browser cache
# Chrome/Firefox: Ctrl+Shift+Delete

# Verify frontend build size
docker exec takehomeassessment-frontend-1 ls -lh /usr/share/nginx/html/

# Rebuild frontend with optimizations
docker compose build --no-cache frontend
docker compose up -d frontend nginx
```

## Recent Updates & New Features

### October 2025 - Performance & Validation Improvements

**Database optimizations:**
- ✅ Enabled WAL (Write-Ahead Logging) mode for better concurrency
- ✅ Configured 64MB cache for faster queries
- ✅ Added connection pooling to prevent lock contention
- ✅ Memory-mapped I/O for large datasets
- ✅ Indexed queries on (ip_address, timestamp)

**Input validation enhancements:**
- ✅ Extracted validation into dedicated package
- ✅ Comprehensive IP address validation (octets 0-255)
- ✅ Timestamp validation with detailed error messages
- ✅ 13 new validation tests for edge cases

**Service comparison fix:**
- ✅ Fixed bug: Services now identified by port+protocol (was port-only)
- ✅ Correctly handles multiple protocols on same port
- ✅ More accurate diff reports

**Type safety improvements:**
- ✅ Removed all `any` types from frontend
- ✅ Added proper TypeScript interfaces
- ✅ Better compile-time error detection

### Still Having Issues?

1. **Gather debug information:**
   ```bash
   # System info
   uname -a
   docker --version
   docker compose version

   # Service status
   docker compose ps

   # Logs
   docker compose logs --tail=100 > logs.txt
   ```

2. **Enable verbose logging:**
   ```bash
   # Backend tests with verbose output
   cd backend
   go test -v ./...

   # E2E tests with debug mode
   ./e2e_test.sh  # Check script output
   ```

3. **Check known issues:**
   - WAL mode files (-wal, -shm) are normal, not errors
   - UNIQUE constraint errors are expected for duplicate uploads
   - Protocol changes now show as removed+added (correct behavior)

4. **File an issue:**
   - Include system information
   - Attach logs from docker compose logs
   - Describe steps to reproduce
   - Expected vs actual behavior

---

**Last Updated:** October 2025
**Covers:** Common issues, performance tuning, recent improvements
**Status:** Comprehensive troubleshooting guide ✅
