package handlers

import (
	"fmt"
	"net/http"

	"github.com/ekristen/dockit/pkg/common"
	"gorm.io/gorm"
)

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
	fmt.Println("here")

	w.WriteHeader(200)
	w.Write([]byte(data))
}
