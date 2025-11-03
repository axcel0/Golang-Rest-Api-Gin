package health

import (
	"context"
	"runtime"
	"syscall"
	"time"

	"gorm.io/gorm"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status  Status                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status     Status                     `json:"status"`
	Timestamp  string                     `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
	System     SystemInfo                 `json:"system"`
}

// SystemInfo represents system-level information
type SystemInfo struct {
	Goroutines    int     `json:"goroutines"`
	MemoryUsedMB  float64 `json:"memory_used_mb"`
	MemoryAllocMB float64 `json:"memory_alloc_mb"`
	GCPauses      uint32  `json:"gc_pauses"`
}

// Checker interface for health checkers
type Checker interface {
	Check(ctx context.Context) ComponentHealth
}

// DatabaseChecker checks database health
type DatabaseChecker struct {
	DB      *gorm.DB
	Timeout time.Duration
}

// Check implements Checker for DatabaseChecker
func (d *DatabaseChecker) Check(ctx context.Context) ComponentHealth {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	sqlDB, err := d.DB.DB()
	if err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "failed to get database instance",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	// Ping database
	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "database ping failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	// Get database stats
	stats := sqlDB.Stats()

	// Check for connection issues
	if stats.OpenConnections == 0 {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: "no open database connections",
			Details: map[string]interface{}{
				"open_connections": stats.OpenConnections,
			},
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "database is responsive",
		Details: map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"max_open":         stats.MaxOpenConnections,
		},
	}
}

// DiskSpaceChecker checks disk space
type DiskSpaceChecker struct {
	Path              string
	WarningThreshold  float64 // percentage (e.g., 80.0 for 80%)
	CriticalThreshold float64 // percentage (e.g., 90.0 for 90%)
}

// Check implements Checker for DiskSpaceChecker
func (d *DiskSpaceChecker) Check(ctx context.Context) ComponentHealth {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(d.Path, &stat); err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "failed to get disk stats",
			Details: map[string]interface{}{
				"error": err.Error(),
				"path":  d.Path,
			},
		}
	}

	// Calculate disk usage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free
	usagePercent := (float64(used) / float64(total)) * 100

	totalGB := float64(total) / (1024 * 1024 * 1024)
	usedGB := float64(used) / (1024 * 1024 * 1024)
	freeGB := float64(free) / (1024 * 1024 * 1024)

	details := map[string]interface{}{
		"path":          d.Path,
		"total_gb":      round(totalGB, 2),
		"used_gb":       round(usedGB, 2),
		"free_gb":       round(freeGB, 2),
		"usage_percent": round(usagePercent, 2),
	}

	// Check thresholds
	if usagePercent >= d.CriticalThreshold {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "critical disk space usage",
			Details: details,
		}
	}

	if usagePercent >= d.WarningThreshold {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: "high disk space usage",
			Details: details,
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "disk space is adequate",
		Details: details,
	}
}

// MemoryChecker checks memory usage
type MemoryChecker struct {
	WarningThresholdMB  float64
	CriticalThresholdMB float64
}

// Check implements Checker for MemoryChecker
func (m *MemoryChecker) Check(ctx context.Context) ComponentHealth {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	allocMB := float64(memStats.Alloc) / (1024 * 1024)
	totalAllocMB := float64(memStats.TotalAlloc) / (1024 * 1024)
	sysMB := float64(memStats.Sys) / (1024 * 1024)

	details := map[string]interface{}{
		"alloc_mb":       round(allocMB, 2),
		"total_alloc_mb": round(totalAllocMB, 2),
		"sys_mb":         round(sysMB, 2),
		"gc_runs":        memStats.NumGC,
		"goroutines":     runtime.NumGoroutine(),
	}

	// Check thresholds
	if allocMB >= m.CriticalThresholdMB {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "critical memory usage",
			Details: details,
		}
	}

	if allocMB >= m.WarningThresholdMB {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: "high memory usage",
			Details: details,
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "memory usage is normal",
		Details: details,
	}
}

// HealthService manages health checks
type HealthService struct {
	checkers map[string]Checker
}

// NewHealthService creates a new health service
func NewHealthService() *HealthService {
	return &HealthService{
		checkers: make(map[string]Checker),
	}
}

// RegisterChecker registers a new health checker
func (s *HealthService) RegisterChecker(name string, checker Checker) {
	s.checkers[name] = checker
}

// CheckHealth performs all health checks and returns the result
func (s *HealthService) CheckHealth(ctx context.Context) HealthResponse {
	components := make(map[string]ComponentHealth)
	overallStatus := StatusHealthy

	// Run all checkers
	for name, checker := range s.checkers {
		health := checker.Check(ctx)
		components[name] = health

		// Determine overall status (worst case wins)
		if health.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if health.Status == StatusDegraded && overallStatus != StatusUnhealthy {
			overallStatus = StatusDegraded
		}
	}

	// Get system info
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now().Format(time.RFC3339),
		Components: components,
		System: SystemInfo{
			Goroutines:    runtime.NumGoroutine(),
			MemoryUsedMB:  round(float64(memStats.Alloc)/(1024*1024), 2),
			MemoryAllocMB: round(float64(memStats.TotalAlloc)/(1024*1024), 2),
			GCPauses:      memStats.NumGC,
		},
	}
}

// Helper function to round float to n decimal places
func round(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}
