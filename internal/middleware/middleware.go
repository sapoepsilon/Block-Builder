package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"docker-management-system/internal/errors"
	"docker-management-system/internal/logging"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := logging.WithRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs request/response details
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		logging.LogRequest(r.Context(), r.Method, r.URL.Path, duration, rw.statusCode)
	})
}

// ErrorHandler handles and formats error responses
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				appErr := &errors.AppError{
					Code:      http.StatusInternalServerError,
					Message:   "Internal server error",
					ErrorType: "server_error",
				}
				logging.LogError(r.Context(), "panic recovered", nil, zap.Any("panic", err))
				respondWithError(w, appErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func respondWithError(w http.ResponseWriter, err *errors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}
