package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

// newResponseWriter creates a new response writer wrapper.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status
	}
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logging returns a middleware that logs HTTP requests.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status
		wrapped := newResponseWriter(w)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log request details
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %s %d bytes",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Round(time.Millisecond),
			wrapped.written,
		)
	})
}

// LoggingWithLogger returns a logging middleware with a custom logger.
func LoggingWithLogger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger.Printf(
				"%s %s %d %s %d bytes",
				r.Method,
				r.URL.Path,
				wrapped.statusCode,
				duration.Round(time.Millisecond),
				wrapped.written,
			)
		})
	}
}

// RequestID adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	var counter uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		reqID := counter
		w.Header().Set("X-Request-ID", formatRequestID(reqID))
		next.ServeHTTP(w, r)
	})
}

// formatRequestID formats the request ID.
func formatRequestID(id uint64) string {
	return "req-" + uitoa(id)
}

// uitoa converts uint64 to string without importing strconv.
func uitoa(val uint64) string {
	if val == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for val > 0 {
		buf[i] = byte('0' + val%10)
		val /= 10
		i--
	}
	return string(buf[i+1:])
}
