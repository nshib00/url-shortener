package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func LoggerMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/logger"),
		)
		log.Info("middleware: enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			wrappedW := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			timeStart := time.Now()
			defer func() {
				entry.Info("middleware: request completed",
					slog.Int("status", wrappedW.Status()),
					slog.Int("bytes", wrappedW.BytesWritten()),
					slog.String("duration", time.Since(timeStart).String()),
				)
			}()

			next.ServeHTTP(wrappedW, r)
		}
		return http.HandlerFunc(fn)
	}
}
