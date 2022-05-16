package handlers

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ekristen/dockit/pkg/db"
)

func (h *handlers) PKICerts(w http.ResponseWriter, r *http.Request) {
	var pki []db.PKI
	sql := h.db.Model(&db.PKI{}).Where("active = ? AND expires_at > ?", true, time.Now().UTC()).Find(&pki)
	if sql.Error != nil {
		w.WriteHeader(500)
		return
	}

	var resData [][]byte

	for _, p := range pki {
		resData = append(resData, []byte(p.X509))
	}

	w.WriteHeader(200)
	w.Write(bytes.Join(resData, []byte("\n")))
}
