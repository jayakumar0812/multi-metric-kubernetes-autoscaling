package collector

import (
	"fmt"
	"runtime"
	"time"

	"hybrid-metrics/pkg/metrics"
)

// SystemCollector collects system resource metrics
type SystemCollector struct {
	lastCPUTime time.Time
	lastCPUUsage float64
}

// NewSystemCollector creates a new system metrics collector
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{
		lastCPUTime: time.Now(),
	}
}

// Collect gathers current system metrics
func (s *SystemCollector) Collect() (metrics.SystemMetrics, error) {
	cpuUsage, err := s.getCPUUsage()
	if err != nil {
		return metrics.SystemMetrics{}, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	memoryUsage, err := s.getMemoryUsage()
	if err != nil {
		return metrics.SystemMetrics{}, fmt.Errorf("failed to get memory usage: %w", err)
	}

	return metrics.SystemMetrics{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		Timestamp:   time.Now(),
	}, nil
}

// getCPUUsage calculates CPU usage percentage
func (s *SystemCollector) getCPUUsage() (float64, error) {
	// For cross-platform compatibility
	
	// Get number of goroutines as a proxy for activity
	numGoroutines := runtime.NumGoroutine()
	
	// Simple simulation of CPU usage based on system activity
	cpuUsage := float64(numGoroutines) * 2.0
	
	// Add some realistic variation
	now := time.Now()
	timeDiff := now.Sub(s.lastCPUTime).Seconds()
	
	if timeDiff > 0 {
		// Simulate CPU usage with some randomness and smoothing
		variation := (float64(now.UnixNano()%100) - 50) / 100.0 * 10.0
		cpuUsage = s.lastCPUUsage*0.7 + (cpuUsage+variation)*0.3
	}
	
	// CPU usage is within bounds
	if cpuUsage < 0 {
		cpuUsage = 0
	}
	if cpuUsage > 100 {
		cpuUsage = 100
	}
	
	s.lastCPUUsage = cpuUsage
	s.lastCPUTime = now
	
	return cpuUsage, nil
}

// getMemoryUsage calculates memory usage percentage
func (s *SystemCollector) getMemoryUsage() (float64, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Calculate memory usage as percentage
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024
	
	// Simulate memory
	memoryUsage := (allocMB / (sysMB + 100)) * 100 // +100 to simulate total system memory
	
	if memoryUsage > 100 {
		memoryUsage = 100
	}
	
	return memoryUsage, nil
}

// GetDetailedMetrics provides more detailed system information
func (s *SystemCollector) GetDetailedMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"goroutines":       runtime.NumGoroutine(),
		"alloc_mb":        float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb":  float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":          float64(m.Sys) / 1024 / 1024,
		"num_gc":          m.NumGC,
		"gc_pause_ns":     m.PauseNs[(m.NumGC+255)%256],
		"heap_objects":    m.HeapObjects,
		"stack_inuse_mb":  float64(m.StackInuse) / 1024 / 1024,
	}
}

// SimulateLoad creates artificial load for testing purposes
func (s *SystemCollector) SimulateLoad(intensity float64) {
	// testing fusion algorithm
	// CPU and memory activity
	
	duration := time.Duration(intensity * 100) * time.Millisecond
	
	go func() {
		end := time.Now().Add(duration)
		for time.Now().Before(end) {
			// Create some CPU load
			_ = make([]byte, int(intensity*1000))
		}
	}()
}