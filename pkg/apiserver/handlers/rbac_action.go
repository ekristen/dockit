package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ekristen/dockit/pkg/apiserver/types"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/db"
	"github.com/ekristen/dockit/pkg/httpauth"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GroupAction struct {
	Name string `json:"name"`
}

func (h *handlers) Action(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(common.ContextReqIDKey)
	log := logrus.WithField("reqID", reqID)

	params := mux.Vars(r)

	auth, err := httpauth.Parse(r)
	if err != nil {
		log.WithError(err).Error("unable to parse auth header")
		w.WriteHeader(500)
		return
	}

	log.WithField("query", r.URL.Query()).Debug("url query")

	log.Debug("basic authentication")
	var user db.User

	sql := h.db.Where("username = ? AND admin = ?", auth.Username(), true).First(&user)
	if sql.Error != nil {
		if sql.Error == gorm.ErrRecordNotFound {
			w.WriteHeader(401)
			return
		}

		log.WithError(sql.Error).Error("unable to query database")
		w.WriteHeader(500)
		return
	}

	if sql.RowsAffected == 0 {
		log.Debug("unknown user")
		w.WriteHeader(401)
		w.Write([]byte("unknown user"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(auth.Password())); err != nil {
		log.WithError(err).Debug("invalid password")
		w.WriteHeader(401)
		w.Write([]byte("invalid password"))
		return
	}

	// we've authenticated successfully ...

	rbac_type, ok := params["rbac_type"]
	if !ok {
		w.WriteHeader(400)
		return
	}
	rbac_entity, ok := params["rbac_entity"]
	if !ok {
		w.WriteHeader(400)
		return
	}

	action, ok := params["action"]
	if !ok {
		w.WriteHeader(400)
		return
	}

	switch rbac_type {
	case "user":
		var user db.User
		sql := h.db.Where("username = ?", rbac_entity).First(&user)
		if sql.Error != nil {
			if sql.Error == gorm.ErrRecordNotFound {
				sendErrorResponse(w, 500, fmt.Errorf("unknown user: %s", rbac_entity))
				return
			}

			logrus.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}

		if sql.RowsAffected == 0 {
			logrus.Warn("no user found")
			w.WriteHeader(400)
			return
		}

		switch action {
		case "enable":
			sql := h.db.Model(&user).Update("active", true)
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}
		case "disable":
			sql := h.db.Model(&user).Update("active", false)
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}
		case "change-password":
			type ChangePassword struct {
				Password string `json:"password"`
			}

			var newPassword ChangePassword

			if err := json.NewDecoder(r.Body).Decode(&newPassword); err != nil {
				logrus.WithError(err).Error("unable to decode json")
			}

			if len(newPassword.Password) < 4 {
				log.WithField(rbac_type, rbac_entity).Info("change password failed due to length")
				w.WriteHeader(400)
				return
			}

			sql := h.db.Model(&user).Update("password", newPassword.Password)
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}

			log.WithField(rbac_type, rbac_entity).Info("change password successful")
		default:
			res := types.Response{
				Success: false,
				Errors: []error{
					fmt.Errorf("unsupported action: %s", action),
				},
			}
			if err := json.NewEncoder(w).Encode(res); err != nil {
				w.WriteHeader(500)
				w.Write([]byte(`{"success": false}`))
			}
		}
	case "group":
		if action == "add" {
			sql := h.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&db.Group{Name: rbac_entity, Active: true})
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}

		}

		var group db.Group
		sql := h.db.Where("name = ?", rbac_entity).First(&group)
		if sql.Error != nil {
			if sql.Error == gorm.ErrRecordNotFound {
				sendErrorResponse(w, 404, fmt.Errorf("unknown group: %s", rbac_entity))
				return
			}

			logrus.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}

		if sql.RowsAffected == 0 {
			logrus.Warn("no group found")
			w.WriteHeader(400)
			return
		}

		switch action {
		case "add":
		case "remove":
			sql := h.db.Model(&group).Delete(&group)
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}

		case "enable":
			sql := h.db.Model(&group).Update("active", true)
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}
		case "disable":
			sql := h.db.Model(&group).Update("active", false)
			if sql.Error != nil {
				logrus.WithError(sql.Error).Error("unable to query database")
				w.WriteHeader(500)
				return
			}
		case "add-member", "remove-member":
			rbac_type2, ok := params["rbac_type_2"]
			if !ok {
				w.WriteHeader(400)
				return
			}
			rbac_entity2, ok := params["rbac_entity_2"]
			if !ok {
				w.WriteHeader(400)
				return
			}

			switch rbac_type2 {
			case "user":
				var user db.User
				sql := h.db.Model(&db.User{}).Where("username = ?", rbac_entity2).First(&user)
				if sql.Error != nil {
					if sql.Error == gorm.ErrRecordNotFound {
						sendErrorResponse(w, 404, fmt.Errorf("unknown user: %s", rbac_entity2))
						return
					}

					logrus.WithError(sql.Error).Error("unable to query database")
					w.WriteHeader(500)
					return
				}

				if r.Method == "PUT" {
					if err := h.db.Model(&group).Association("Users").Append(&user); err != nil {
						if sql.Error == gorm.ErrRecordNotFound {
							logrus.WithError(sql.Error).Warnf("group not found: %s", rbac_entity)
							sendErrorResponse(w, 404, fmt.Errorf("unknown group: %s", rbac_entity))
							return
						}
						logrus.WithError(sql.Error).Error("unable to query database")
						w.WriteHeader(500)
						return
					}
				} else if r.Method == "DELETE" {
					if err := h.db.Model(&group).Association("Users").Delete(&user); err != nil {
						if sql.Error == gorm.ErrRecordNotFound {
							logrus.WithError(sql.Error).Warnf("group not found: %s", rbac_entity)
							sendErrorResponse(w, 404, fmt.Errorf("unknown group: %s", rbac_entity))
							return
						}

						logrus.WithError(sql.Error).Error("unable to query database")
						w.WriteHeader(500)
						return
					}
				}
			default:
				logrus.Error("unsupported rbac type: %s", rbac_type2)
				w.WriteHeader(501)
				return
			}
		case "permissions":
			w.WriteHeader(200)
		default:
			log.Errorf("unknown action: %s", action)
			w.WriteHeader(501)
		}
	default:
		log.Errorf("unknown rbac type: %s", rbac_type)
		w.WriteHeader(501)
		return
	}

	w.WriteHeader(200)
	w.Header().Add("content-type", "application/json")
	w.Write([]byte(`{"success": true}`))
}
