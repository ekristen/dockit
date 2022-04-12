package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ekristen/dockit/pkg/apiserver/response"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/db"
	"github.com/ekristen/dockit/pkg/httpauth"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (h *handlers) Permission(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(common.ContextReqIDKey)
	log := logrus.WithField("reqID", reqID)

	params := mux.Vars(r)

	res := response.New(w, r)

	auth, err := httpauth.Parse(r)
	if err != nil {
		log.WithError(err).Error("unable to parse auth header")
		res.AddError(err).Send(500)
		return
	}

	log.WithField("query", r.URL.Query()).Debug("url query")

	log.Debug("basic authentication")
	var user db.User

	sql := h.db.Where("username = ? AND admin = ?", auth.Username(), true).First(&user)
	if sql.Error != nil {
		if sql.Error == gorm.ErrRecordNotFound {
			err := fmt.Errorf("unauthorized")
			res.AddError(err).Send(401)
			return
		}

		log.WithError(sql.Error).Error("unable to query database")
		res.AddError(DBError).Send(500)
		return
	}

	if sql.RowsAffected == 0 {
		log.Debug("unknown user")
		res.AddError(UnauthorizedError).Send(401)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(auth.Password())); err != nil {
		log.WithError(err).Debug("invalid password")
		res.AddError(UnauthorizedError).Send(401)
		return
	}

	// we've authenticated successfully ...

	var entityID int64

	rbac_type, ok := params["rbac_type"]
	if !ok {
		log.Errorf("missing rbac_type parameter")
		res.AddError(errors.New("missing rbac_type parameter")).Send(422)
		return
	}
	rbac_entity, ok := params["rbac_entity"]
	if !ok {
		log.Errorf("missing rbac_entity parameter")
		res.AddError(errors.New("missing rbac_entity parameter")).Send(422)
		return
	}

	switch rbac_type {
	case "user":
		var user db.User
		sql := h.db.Where("username = ?", rbac_entity).First(&user)
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			res.AddError(DBError).Send(500)
			return
		}

		if sql.RowsAffected == 0 {
			logrus.Warn("no user found")
			w.WriteHeader(400)
			return
		}
		entityID = user.ID
	case "group":
		var group db.Group
		sql := h.db.Where("name = ?", rbac_entity).First(&group)
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
	default:
		logrus.Errorf("unknown rbac type: %s", rbac_type)
		w.WriteHeader(500)
		return
	}

	switch r.Method {
	case "PUT":
		sql = h.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "type"}, {Name: "name"}, {Name: "class"}},
			DoUpdates: clause.AssignmentColumns([]string{"action"}),
		}).Create(&db.Permission{
			Type:     db.PermissionType(params["type"]),
			Name:     params["name"],
			Action:   db.PermissionAction(params["action"]),
			EntityID: entityID,
		})
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			res.AddError(DBError).Send(500)
			return
		}
	case "DELETE":
		sql = h.db.Model(&db.Permission{}).
			Where("type = ?", params["type"]).
			Where("name = ?", params["name"]).
			Where("action = ?", params["action"]).
			Where("entity_id = ?", entityID).
			Delete(&db.Permission{})
		if sql.Error != nil {
			logrus.WithError(sql.Error).Error("unable to query database")
			res.AddError(DBError).Send(500)
			return
		}
	}

	res.Success().Send(200)
}
