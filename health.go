package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a specific component
type ComponentHealth struct {
	Name        string        `json:"name"`
	Status      HealthStatus  `json:"status"`
	LastChecked time.Time     `json:"last_checked"`
	Duration    time.Duration `json:"duration_ms"`
	Message     string        `json:"message,omitempty"`
	Details     interface{}   `json:"details,omitempty"`
}

// HealthCheck represents the overall health check response
type HealthCheck struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     time.Duration              `json:"uptime_seconds"`
	Components map[string]ComponentHealth `json:"components"`
	System     SystemHealth               `json:"system"`
}

// SystemHealth represents system-level health metrics
type SystemHealth struct {
	Memory     MemoryStats `json:"memory"`
	Goroutines int         `json:"goroutines"`
	GOMAXPROCS int         `json:"gomaxprocs"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc_bytes"`
	TotalAlloc uint64 `json:"total_alloc_bytes"`
	Sys        uint64 `json:"sys_bytes"`
	NumGC      uint32 `json:"num_gc"`
}

// HealthChecker performs health checks on various components
type HealthChecker struct {
	startTime      time.Time
	fileOps        *RobustFileOperations
	circuitBreaker *CircuitBreaker
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		fileOps:   NewRobustFileOperations(),
	}
}

// SetCircuitBreaker sets the circuit breaker for health monitoring
func (hc *HealthChecker) SetCircuitBreaker(cb *CircuitBreaker) {
	hc.circuitBreaker = cb
}

// CheckHealth performs a comprehensive health check
func (hc *HealthChecker) CheckHealth(requestID string) HealthCheck {
	components := make(map[string]ComponentHealth)

	// Check file system health
	components["filesystem"] = hc.checkFileSystemHealth(requestID)

	// Check OpenAI connectivity (if circuit breaker is available)
	if hc.circuitBreaker != nil {
		components["openai"] = hc.checkOpenAIHealth(requestID)
	}

	// Check memory health
	components["memory"] = hc.checkMemoryHealth(requestID)

	// Check disk space
	components["disk_space"] = hc.checkDiskSpaceHealth(requestID)

	// Determine overall status
	overallStatus := hc.determineOverallStatus(components)

	// Get system stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return HealthCheck{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    "1.0.0", // This could be injected from build info
		Uptime:     time.Since(hc.startTime),
		Components: components,
		System: SystemHealth{
			Memory: MemoryStats{
				Alloc:      memStats.Alloc,
				TotalAlloc: memStats.TotalAlloc,
				Sys:        memStats.Sys,
				NumGC:      memStats.NumGC,
			},
			Goroutines: runtime.NumGoroutine(),
			GOMAXPROCS: runtime.GOMAXPROCS(0),
		},
	}
}

// checkFileSystemHealth checks if the file system is accessible
func (hc *HealthChecker) checkFileSystemHealth(requestID string) ComponentHealth {
	start := time.Now()

	// Try to validate write access to data directory
	err := hc.fileOps.ValidateWriteAccess("./", requestID)
	duration := time.Since(start)

	if err != nil {
		logger.Error(fmt.Sprintf("[%s] File system health check failed: %v", requestID, err))
		return ComponentHealth{
			Name:        "filesystem",
			Status:      HealthStatusUnhealthy,
			LastChecked: time.Now(),
			Duration:    duration,
			Message:     fmt.Sprintf("File system write access failed: %v", err),
		}
	}

	return ComponentHealth{
		Name:        "filesystem",
		Status:      HealthStatusHealthy,
		LastChecked: time.Now(),
		Duration:    duration,
		Message:     "File system accessible",
	}
}

// checkOpenAIHealth checks OpenAI API connectivity status
func (hc *HealthChecker) checkOpenAIHealth(requestID string) ComponentHealth {
	_ = requestID // delint
	start := time.Now()

	metrics := hc.circuitBreaker.GetMetrics()
	duration := time.Since(start)

	var status HealthStatus
	var message string

	switch metrics.State {
	case CircuitClosed:
		status = HealthStatusHealthy
		message = "OpenAI circuit breaker closed (healthy)"
	case CircuitHalfOpen:
		status = HealthStatusDegraded
		message = "OpenAI circuit breaker half-open (recovering)"
	case CircuitOpen:
		status = HealthStatusUnhealthy
		message = "OpenAI circuit breaker open (failing)"
	default:
		status = HealthStatusUnhealthy
		message = "OpenAI circuit breaker in unknown state"
	}

	return ComponentHealth{
		Name:        "openai",
		Status:      status,
		LastChecked: time.Now(),
		Duration:    duration,
		Message:     message,
		Details: map[string]interface{}{
			"state":        metrics.State.String(),
			"failures":     metrics.TotalFailures,
			"successes":    metrics.TotalSuccesses,
			"last_failure": metrics.LastFailureTime,
		},
	}
}

// checkMemoryHealth checks memory usage
func (hc *HealthChecker) checkMemoryHealth(requestID string) ComponentHealth {
	_ = requestID // delint
	start := time.Now()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	duration := time.Since(start)

	// Check if memory usage is concerning (>90% of available)
	allocMB := memStats.Alloc / 1024 / 1024
	sysMB := memStats.Sys / 1024 / 1024

	var status HealthStatus
	var message string

	if allocMB > 1000 { // More than 1GB allocated
		status = HealthStatusDegraded
		message = fmt.Sprintf("High memory usage: %d MB allocated", allocMB)
	} else if allocMB > 2000 { // More than 2GB allocated
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Very high memory usage: %d MB allocated", allocMB)
	} else {
		status = HealthStatusHealthy
		message = fmt.Sprintf("Memory usage normal: %d MB allocated", allocMB)
	}

	return ComponentHealth{
		Name:        "memory",
		Status:      status,
		LastChecked: time.Now(),
		Duration:    duration,
		Message:     message,
		Details: map[string]interface{}{
			"alloc_mb":   allocMB,
			"sys_mb":     sysMB,
			"num_gc":     memStats.NumGC,
			"goroutines": runtime.NumGoroutine(),
		},
	}
}

// checkDiskSpaceHealth checks available disk space
func (hc *HealthChecker) checkDiskSpaceHealth(requestID string) ComponentHealth {
	start := time.Now()

	diskInfo, err := hc.fileOps.CheckDiskSpace("./", requestID)
	duration := time.Since(start)

	if err != nil {
		return ComponentHealth{
			Name:        "disk_space",
			Status:      HealthStatusUnhealthy,
			LastChecked: time.Now(),
			Duration:    duration,
			Message:     fmt.Sprintf("Failed to check disk space: %v", err),
		}
	}

	freeSpaceGB := diskInfo.Available / (1024 * 1024 * 1024)
	totalSpaceGB := diskInfo.Total / (1024 * 1024 * 1024)
	usagePercent := diskInfo.UsedPercent

	var status HealthStatus
	var message string

	if usagePercent > 95 {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Critical disk space: %.1f%% used", usagePercent)
	} else if usagePercent > 85 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Low disk space: %.1f%% used", usagePercent)
	} else {
		status = HealthStatusHealthy
		message = fmt.Sprintf("Disk space healthy: %.1f%% used", usagePercent)
	}

	return ComponentHealth{
		Name:        "disk_space",
		Status:      status,
		LastChecked: time.Now(),
		Duration:    duration,
		Message:     message,
		Details: map[string]interface{}{
			"free_gb":       freeSpaceGB,
			"total_gb":      totalSpaceGB,
			"usage_percent": usagePercent,
		},
	}
}

// determineOverallStatus determines the overall health status based on components
func (hc *HealthChecker) determineOverallStatus(components map[string]ComponentHealth) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, component := range components {
		switch component.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// Global health checker
var globalHealthChecker = NewHealthChecker()

// GetHealthChecker returns the global health checker
func GetHealthChecker() *HealthChecker {
	return globalHealthChecker
}
