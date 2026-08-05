package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type loggingResponceWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponceWriter) WriteHeader(code int) { //Переопределение метода WriteHeader
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lwr := &loggingResponceWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lwr, r)

		duration := time.Since(start)

		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		logger.Info("HTTP request processed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", lwr.statusCode),
			slog.Duration("duration", duration),
		)

	})
}
