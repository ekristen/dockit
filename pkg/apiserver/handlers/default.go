package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ekristen/dockit/pkg/apiserver/types"
	"github.com/ekristen/dockit/pkg/common"
	"gorm.io/gorm"
)

var DBError = errors.New("database error")
var UnauthorizedError = errors.New("unauthorized")

type handlers struct {
	db *gorm.DB
}

func New(db *gorm.DB) *handlers {
	return &handlers{
		db: db,
	}
}

func (h *handlers) Root(w http.ResponseWriter, r *http.Request) {
	data := fmt.Sprintf(`{"name":"%s","version":"%s"}`, common.AppVersion.Name, common.AppVersion.Summary)

	w.WriteHeader(200)
	w.Write([]byte(data))
}

func sendErrorResponse(w http.ResponseWriter, code int, err error) {
	res := types.Response{
		Success: false,
		Errors:  types.Errors{err},
	}
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`{"success": false}`))
	}
}
