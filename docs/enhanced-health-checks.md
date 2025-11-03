# Enhanced Health Checks

Production-ready health check system with component-level monitoring for databases, disk space, and memory usage.

## Overview

The application provides comprehensive health monitoring through the `/health` endpoint, which checks multiple system components and returns detailed status information. This enables:

- **Kubernetes liveness probes** - Detect unhealthy pods
- **Load balancer health checks** - Route traffic only to healthy instances
- **Monitoring dashboards** - Track system health metrics
- **Incident response** - Quick diagnosis of component failures

## Endpoints

### 1. Enhanced Health Check

**Endpoint:** `GET /health`

**Description:** Comprehensive health check with component-level details

**Response Codes:**
- `200 OK` - All components healthy or degraded
- `503 Service Unavailable` - One or more components unhealthy

**Response Structure:**

```json
{
  "status": "healthy|degraded|unhealthy",
  "timestamp": "2025-10-31T16:00:39+07:00",
  "components": {
    "database": { ... },
    "disk": { ... },
    "memory": { ... }
  },
  "system": {
    "goroutines": 4,
    "memory_used_mb": 9.89,
    "memory_alloc_mb": 10.54,
    "gc_pauses": 2
  }
}
```

---

### 2. Readiness Check

**Endpoint:** `GET /ready`

**Description:** Lightweight readiness probe for Kubernetes/Docker

**Response:**
```json
{
  "status": "ready",
  "service": "Go-Lang-project-01"
}
```

**Use Case:** Fast endpoint for load balancer health checks that don't need detailed component information.

---

## Component Checks

### Database Health

**Component Name:** `database`

**Checks:**
- Database connection availability
- Ping response time (5 second timeout)
- Connection pool statistics

**Status Conditions:**

| Status | Condition | Description |
|--------|-----------|-------------|
| **Healthy** | Ping successful, connections > 0 | Database is responsive |
| **Degraded** | Ping successful, connections = 0 | No active connections (unusual) |
| **Unhealthy** | Ping failed or timeout | Database unreachable |

**Response Example:**

```json
{
  "database": {
    "status": "healthy",
    "message": "database is responsive",
    "details": {
      "open_connections": 1,
      "in_use": 0,
      "idle": 1,
      "max_open": 0
    }
  }
}
```

**Configuration:**
```go
healthService.RegisterChecker("database", &health.DatabaseChecker{
    DB:      db,
    Timeout: 5 * time.Second,  // Ping timeout
})
```

---

### Disk Space Health

**Component Name:** `disk`

**Checks:**
- Disk usage percentage
- Available space in GB
- Path: `/` (root filesystem)

**Status Conditions:**

| Status | Condition | Description |
|--------|-----------|-------------|
| **Healthy** | Usage < 80% | Adequate disk space |
| **Degraded** | Usage 80-89% | High disk usage, investigate |
| **Unhealthy** | Usage ≥ 90% | Critical disk space, immediate action needed |

**Response Example:**

```json
{
  "disk": {
    "status": "healthy",
    "message": "disk space is adequate",
    "details": {
      "path": "/",
      "total_gb": 468.09,
      "used_gb": 125.09,
      "free_gb": 343.0,
      "usage_percent": 26.72
    }
  }
}
```

**Configuration:**
```go
healthService.RegisterChecker("disk", &health.DiskSpaceChecker{
    Path:              "/",
    WarningThreshold:  80.0,  // % usage for degraded
    CriticalThreshold: 90.0,  // % usage for unhealthy
})
```

---

### Memory Health

**Component Name:** `memory`

**Checks:**
- Allocated memory in MB
- Total allocated memory (cumulative)
- System memory reserved
- Garbage collection runs
- Active goroutines

**Status Conditions:**

| Status | Condition | Description |
|--------|-----------|-------------|
| **Healthy** | Alloc < 500MB | Normal memory usage |
| **Degraded** | Alloc 500-1023MB | High memory usage, monitor |
| **Unhealthy** | Alloc ≥ 1GB | Critical memory usage, possible leak |

**Response Example:**

```json
{
  "memory": {
    "status": "healthy",
    "message": "memory usage is normal",
    "details": {
      "alloc_mb": 9.89,
      "total_alloc_mb": 10.54,
      "sys_mb": 22.08,
      "gc_runs": 2,
      "goroutines": 4
    }
  }
}
```

**Configuration:**
```go
healthService.RegisterChecker("memory", &health.MemoryChecker{
    WarningThresholdMB:  500,   // MB for degraded
    CriticalThresholdMB: 1024,  // MB for unhealthy
})
```

---

## Status Levels

### Healthy ✅

**Definition:** All components functioning normally

**HTTP Status:** `200 OK`

**Action:** None required

**Example:**
```json
{
  "status": "healthy",
  "components": {
    "database": { "status": "healthy", ... },
    "disk": { "status": "healthy", ... },
    "memory": { "status": "healthy", ... }
  }
}
```

---

### Degraded ⚠️

**Definition:** Service operational but one or more components need attention

**HTTP Status:** `200 OK` (still accepting traffic)

**Action:** Investigate and monitor

**Triggers:**
- Disk usage 80-89%
- Memory usage 500-1023MB
- Database has zero connections (unusual but not critical)

**Example:**
```json
{
  "status": "degraded",
  "components": {
    "database": { "status": "healthy", ... },
    "disk": { "status": "degraded", "message": "high disk space usage", ... },
    "memory": { "status": "healthy", ... }
  }
}
```

---

### Unhealthy ❌

**Definition:** Critical failure, service should not receive traffic

**HTTP Status:** `503 Service Unavailable`

**Action:** Immediate intervention required

**Triggers:**
- Database unreachable or ping timeout
- Disk usage ≥ 90%
- Memory usage ≥ 1GB

**Example:**
```json
{
  "status": "unhealthy",
  "components": {
    "database": { 
      "status": "unhealthy", 
      "message": "database ping failed",
      "details": { "error": "connection refused" }
    },
    "disk": { "status": "healthy", ... },
    "memory": { "status": "healthy", ... }
  }
}
```

---

## Integration

### Kubernetes Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

**Behavior:**
- Kubernetes checks `/health` every 10 seconds
- If 3 consecutive failures (503 status), pod is restarted
- Handles unhealthy components (DB down, disk full, memory leak)

---

### Kubernetes Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 2
```

**Behavior:**
- Kubernetes checks `/ready` every 5 seconds
- If 2 consecutive failures, removes pod from service endpoints
- Prevents traffic routing to starting/stopping pods

---

### Docker Compose Health Check

```yaml
services:
  api:
    image: go-rest-api:latest
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 40s
```

---

### Load Balancer Health Check

**AWS ALB:**
```hcl
resource "aws_lb_target_group" "api" {
  health_check {
    path                = "/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }
}
```

**NGINX:**
```nginx
upstream api_backend {
    server api1:8080 max_fails=3 fail_timeout=30s;
    server api2:8080 max_fails=3 fail_timeout=30s;
}

location /health {
    proxy_pass http://api_backend;
    proxy_connect_timeout 5s;
}
```

---

## Monitoring & Alerting

### Prometheus Metrics

Health checks automatically generate Prometheus metrics:

```promql
# Health check failures
sum(rate(http_requests_total{endpoint="/health",status="503"}[5m]))

# Health check latency
histogram_quantile(0.95, 
  rate(http_request_duration_seconds_bucket{endpoint="/health"}[5m])
)
```

### Alert Rules

```yaml
groups:
  - name: health_checks
    rules:
      - alert: ServiceUnhealthy
        expr: |
          sum(rate(http_requests_total{endpoint="/health",status="503"}[5m])) > 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Service reporting unhealthy status"
          
      - alert: HealthCheckSlow
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket{endpoint="/health"}[5m])
          ) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Health checks taking too long"
```

---

## Testing

### Manual Testing

```bash
# Test healthy status
curl http://localhost:8080/health | jq .

# Test specific component
curl http://localhost:8080/health | jq '.components.database'

# Check status code
curl -I http://localhost:8080/health

# Test readiness
curl http://localhost:8080/ready
```

### Load Testing

```bash
# Generate load and monitor health
while true; do
  curl -s http://localhost:8080/health | jq -r '.status'
  sleep 5
done
```

### Simulate Failures

**Database failure:**
```bash
# Stop MySQL/PostgreSQL
sudo systemctl stop mysql

# Check health (should be unhealthy)
curl http://localhost:8080/health | jq '.components.database'
```

**Disk space pressure:**
```bash
# Fill disk (be careful!)
dd if=/dev/zero of=/tmp/fillfile bs=1M count=10000

# Check health (may show degraded/unhealthy)
curl http://localhost:8080/health | jq '.components.disk'
```

---

## Implementation Details

### Architecture

```
┌─────────────────────────────────────────┐
│         HealthHandler                    │
│  (HTTP Request Handler)                  │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│         HealthService                    │
│  (Orchestrates all checkers)             │
└────────────────┬────────────────────────┘
                 │
        ┌────────┼────────┐
        ▼        ▼        ▼
    ┌──────┐ ┌──────┐ ┌────────┐
    │  DB  │ │ Disk │ │ Memory │
    │Check │ │Check │ │ Check  │
    └──────┘ └──────┘ └────────┘
```

### Health Service

**Location:** `internal/health/health.go`

**Key Components:**
- `Checker` interface - Contract for all health checkers
- `HealthService` - Orchestrates multiple checkers
- `DatabaseChecker` - Checks database connectivity
- `DiskSpaceChecker` - Monitors disk usage
- `MemoryChecker` - Tracks memory consumption

**Registration:**
```go
healthService := health.NewHealthService()

healthService.RegisterChecker("database", &health.DatabaseChecker{
    DB: db, Timeout: 5*time.Second,
})

healthService.RegisterChecker("disk", &health.DiskSpaceChecker{
    Path: "/", WarningThreshold: 80.0, CriticalThreshold: 90.0,
})

healthService.RegisterChecker("memory", &health.MemoryChecker{
    WarningThresholdMB: 500, CriticalThresholdMB: 1024,
})
```

---

## Performance Considerations

### Response Time

- **Target:** < 100ms for healthy checks
- **Timeout:** 5 seconds for database ping
- **Typical:** 10-50ms when all healthy

**Optimization:**
- Checkers run concurrently (can be parallelized)
- Database timeout prevents hanging
- Syscall for disk stats is fast (<1ms)
- Memory stats from runtime (no syscall)

### Resource Impact

- **CPU:** Negligible (<0.1% per check)
- **Memory:** ~100KB for health service
- **Network:** Database ping only

### Caching

Health checks are **NOT cached** by default to provide real-time status. If needed, implement caching:

```go
// Cache health results for 5 seconds
var (
    cachedHealth *health.HealthResponse
    cacheExpiry  time.Time
    cacheMutex   sync.RWMutex
)

func getCachedHealth() *health.HealthResponse {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    
    if time.Now().Before(cacheExpiry) {
        return cachedHealth
    }
    return nil
}
```

---

## Customization

### Adding Custom Checkers

Implement the `Checker` interface:

```go
type RedisChecker struct {
    Client *redis.Client
}

func (r *RedisChecker) Check(ctx context.Context) health.ComponentHealth {
    _, err := r.Client.Ping(ctx).Result()
    if err != nil {
        return health.ComponentHealth{
            Status: health.StatusUnhealthy,
            Message: "redis ping failed",
            Details: map[string]interface{}{"error": err.Error()},
        }
    }
    
    return health.ComponentHealth{
        Status: health.StatusHealthy,
        Message: "redis is responsive",
    }
}
```

Register:
```go
healthService.RegisterChecker("redis", &RedisChecker{Client: redisClient})
```

---

## Troubleshooting

### Health Check Returns 503

**Possible Causes:**
1. Database unreachable
2. Disk usage ≥ 90%
3. Memory usage ≥ 1GB

**Debug:**
```bash
# Check which component is unhealthy
curl http://localhost:8080/health | jq '.components | to_entries[] | select(.value.status == "unhealthy")'

# View error details
curl http://localhost:8080/health | jq '.components.database.details.error'
```

### Slow Health Checks

**Symptoms:** Health check takes > 1 second

**Causes:**
- Database ping timeout
- Slow disk I/O

**Solution:**
```bash
# Check health check duration
curl -w "\nTime: %{time_total}s\n" http://localhost:8080/health

# Reduce database timeout in code
Timeout: 2 * time.Second  // Instead of 5
```

---

## Summary

✅ **Implemented:**
- Enhanced health check service
- Component-level monitoring (DB, disk, memory)
- Three-tier status system (healthy/degraded/unhealthy)
- Production-ready Kubernetes probes
- Comprehensive system metrics
- Detailed error reporting

📊 **Checks Performed:**
- Database connectivity and pool stats
- Disk space usage with thresholds
- Memory allocation and GC stats
- Goroutine count
- System-level metrics

🎯 **Ready for:**
- Kubernetes liveness/readiness probes
- Load balancer health checks
- Monitoring dashboard integration
- Automated alerting
- Production deployment

**Status:** ✅ COMPLETE - Enhanced health checks fully operational
**Date:** 2025-10-31
**Version:** 2.0.0
