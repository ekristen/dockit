package handlers

import (
	"net/http"

	"github.com/ekristen/dockit/pkg/db"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

func (h *handlers) Grant(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var entityID int64

	if v, ok := params["user"]; ok {
		var user db.User
		sql := h.db.Where("username = ?", v).First(&user)
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}

		if sql.RowsAffected == 0 {
			logrus.Warn("no user found")
			w.WriteHeader(400)
			return
		}
		entityID = user.ID
	} else if v, ok := params["group"]; ok {
		var group db.Group
		sql := h.db.Where("name = ?", v).First(&group)
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}

		if sql.RowsAffected == 0 {
			logrus.Warn("no group found")
			w.WriteHeader(400)
			return
		}

		entityID = group.ID
	} else {
		w.WriteHeader(500)
		return
	}

	sql := h.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&db.Permission{
			Type:     db.PermissionType(params["type"]),
			Name:     params["name"],
			Action:   db.PermissionAction(params["action"]),
			EntityID: entityID,
		})
	if sql.Error != nil {
		logrus.WithError(sql.Error).Error("unable to query database")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (h *handlers) Revoke(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var entityID int64

	if v, ok := params["user"]; ok {
		var user db.User
		sql := h.db.Where("username = ?", v).First(&user)
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}
		entityID = user.ID
	} else if v, ok := params["group"]; ok {
		var group db.Group
		sql := h.db.Where("name = ?", v).First(&group)
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}
		entityID = group.ID
	} else {
		w.WriteHeader(500)
		return
	}

	sql := h.db.
		Where("type = ?", params["type"]).
		Where("name = ?", params["name"]).
		Where("action = ?", params["action"]).
		Where("entity_id = ?", entityID).
		Delete(&db.Permission{})
	if sql.Error != nil {
		logrus.WithError(sql.Error).Error("unable to query database")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}
