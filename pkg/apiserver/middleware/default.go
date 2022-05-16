package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/sirupsen/logrus"
)

type middleware struct {
	log *logrus.Entry
}

func NewToken(log *logrus.Entry) *middleware {
	return &middleware{
		log: log.WithField("component", "middleware"),
	}
}

// Middleware function, which will be called for each request
func (m *middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id string

		if reqID := r.Header.Get("X-Request-Id"); reqID != "" {
			id = reqID
		} else {
			id = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), common.ContextReqIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
