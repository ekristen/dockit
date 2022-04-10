package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ekristen/dockit/pkg/common"
	"github.com/sirupsen/logrus"
)

// realIP get the real IP from http request
func realIP(req *http.Request) string {
	ra := req.RemoteAddr
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func LoggingMiddleware2(logger *logrus.Entry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Context().Value(common.ContextReqIDKey)

			if reqID != nil {
				logger = logger.WithField("reqID", reqID)
			}

			if remoteAddr := realIP(r); remoteAddr != "" {
				logger = logger.WithField("remoteAddr", remoteAddr)
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.WithError(err.(error)).WithField("status", http.StatusInternalServerError).Error("recovered error")
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)

			// Per Shannon Huff 2021-11-20, only log requests if it's not
			// a health check request
			if !strings.Contains(r.URL.EscapedPath(), "healthz") {
				logger.WithFields(logrus.Fields{
					"status":   wrapped.status,
					"method":   r.Method,
					"path":     r.URL.EscapedPath(),
					"duration": time.Since(start),
				}).Infof("handled: %d", wrapped.status)
			}
		}

		return http.HandlerFunc(fn)
	}
}
