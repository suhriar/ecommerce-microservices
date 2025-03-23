package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ResponseWriter adalah wrapper untuk http.ResponseWriter yang menyimpan status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader menangkap status code dari response
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware adalah middleware untuk mencatat log setiap request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter untuk mendapatkan status code
		responseWriter := &ResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Menyimpan logger dalam request context dengan request ID
		logger := log.With().
			Str("request_id", uuid.New().String()).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Logger()

		ctx := logger.WithContext(r.Context())
		r = r.WithContext(ctx)

		// Panggil handler berikutnya
		next.ServeHTTP(responseWriter, r)

		// Log setelah request selesai
		responseTime := time.Since(start)

		// Tentukan level logging berdasarkan status code
		logEvent := logger.Info()
		if responseWriter.statusCode >= 400 && responseWriter.statusCode < 500 {
			logEvent = logger.Warn()
		} else if responseWriter.statusCode >= 500 {
			logEvent = logger.Error()
		}

		logEvent.
			Int("status", responseWriter.statusCode).
			Dur("duration_ms", responseTime).
			Msg("request completed")
	})
}
