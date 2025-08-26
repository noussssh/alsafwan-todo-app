package middleware

import (
	"context"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics tracks request performance metrics
type PerformanceMetrics struct {
	RequestCount     int64
	TotalDuration    time.Duration
	AverageResponse  time.Duration
	SlowRequestCount int64 // Requests taking >1 second
	ErrorCount       int64
}

var metrics = &PerformanceMetrics{}

// PerformanceLogger logs request performance metrics
func PerformanceLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		
		// Update metrics (in production, use atomic operations for thread safety)
		metrics.RequestCount++
		metrics.TotalDuration += duration
		metrics.AverageResponse = metrics.TotalDuration / time.Duration(metrics.RequestCount)
		
		if duration > time.Second {
			metrics.SlowRequestCount++
		}
		
		if statusCode >= 400 {
			metrics.ErrorCount++
		}
		
		// Log slow requests
		if duration > 500*time.Millisecond {
			log.Printf("SLOW REQUEST: %s %s took %v (status: %d)", 
				c.Request.Method, c.Request.URL.Path, duration, statusCode)
		}
		
		// Add performance headers for monitoring
		c.Header("X-Response-Time", strconv.FormatInt(duration.Nanoseconds()/1000000, 10)+"ms")
		c.Header("X-Request-ID", generateRequestID())
	}
}

// HealthCheck provides a health check endpoint with performance metrics
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"metrics": gin.H{
				"total_requests":     metrics.RequestCount,
				"average_response":   metrics.AverageResponse.String(),
				"slow_requests":      metrics.SlowRequestCount,
				"error_count":        metrics.ErrorCount,
				"slow_request_ratio": float64(metrics.SlowRequestCount) / float64(metrics.RequestCount),
			},
			"memory": getMemoryUsage(),
		})
	}
}

// RequestSizeLimit limits request body size for security and performance
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(413, gin.H{"error": "Request too large"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TimeoutMiddleware adds timeout to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		
		// Replace request context
		c.Request = c.Request.WithContext(timeoutCtx)
		
		c.Next()
	}
}

// generateRequestID creates a simple request ID for tracing
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// getMemoryUsage returns basic memory statistics
func getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"alloc_mb":      bToMb(m.Alloc),
		"total_alloc_mb": bToMb(m.TotalAlloc),
		"sys_mb":        bToMb(m.Sys),
		"num_gc":        m.NumGC,
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}