package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed         = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression     = gzip.NoCompression
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(io.Discard, DefaultCompression)
		return gz
	},
}

// Gzip returns a middleware to enable gzip compression for responses
func Gzip(level int) gin.HandlerFunc {
	return GzipWithConfig(GzipConfig{
		Level: level,
	})
}

type GzipConfig struct {
	Level int
	// Skip compression for specific paths
	ExcludedPaths []string
	// Skip compression for specific extensions
	ExcludedExtensions []string
	// Skip compression for specific content types
	ExcludedContentTypes []string
}

func GzipWithConfig(config GzipConfig) gin.HandlerFunc {
	if config.Level == 0 {
		config.Level = DefaultCompression
	}

	// Set default excluded extensions (already compressed)
	if len(config.ExcludedExtensions) == 0 {
		config.ExcludedExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".pdf", ".zip", ".gz", ".mp4", ".avi", ".mov", ".woff", ".woff2", ".ttf", ".eot"}
	}

	// Set default excluded content types
	if len(config.ExcludedContentTypes) == 0 {
		config.ExcludedContentTypes = []string{"image/", "video/", "audio/", "application/pdf", "application/zip", "application/gzip"}
	}

	return func(c *gin.Context) {
		// Skip if client doesn't accept gzip
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Skip excluded paths
		for _, path := range config.ExcludedPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Skip excluded extensions
		for _, ext := range config.ExcludedExtensions {
			if strings.HasSuffix(c.Request.URL.Path, ext) {
				c.Next()
				return
			}
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		gz.Reset(c.Writer)
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Wrap the response writer
		c.Writer = &gzipWriter{c.Writer, gz}

		c.Next()

		// Check if we should skip compression based on content type
		contentType := c.Writer.Header().Get("Content-Type")
		for _, excludedType := range config.ExcludedContentTypes {
			if strings.Contains(contentType, excludedType) {
				return
			}
		}
	}
}

// StaticFileHeaders adds appropriate cache and compression headers for static files
func StaticFileHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Set appropriate cache headers based on file type
		if strings.HasPrefix(path, "/static/") {
			// Static assets can be cached for longer
			if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") {
				c.Header("Cache-Control", "public, max-age=31536000") // 1 year
				c.Header("Expires", "Thu, 31 Dec 2025 23:55:55 GMT")
			} else if strings.HasSuffix(path, ".ico") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".gif") || strings.HasSuffix(path, ".svg") {
				c.Header("Cache-Control", "public, max-age=604800") // 1 week
			}
			
			// Add security headers
			c.Header("X-Content-Type-Options", "nosniff")
		}

		c.Next()
	}
}