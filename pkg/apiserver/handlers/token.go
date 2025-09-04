package handlers

import (
	"crypto"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ekristen/dockit/pkg/apiserver/response"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/db"
	"github.com/ekristen/dockit/pkg/docker"
	"github.com/ekristen/dockit/pkg/httpauth"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TokenResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type TokenClaims struct {
	Access []docker.Scope `json:"access,omitempty"`
	jwt.StandardClaims
}

func (h *handlers) BearerToken(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}

func (h *handlers) Token(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(common.ContextReqIDKey)
	log := logrus.WithField("reqID", reqID)

	auth, err := httpauth.Parse(r)
	if err != nil {
		log.WithError(err).Error("unable to parse auth header")
		w.WriteHeader(500)
		return
	}

	var audience string = ""
	var subject string = ""
	var scopes []docker.Scope

	log.WithField("query", r.URL.Query()).Debug("url query")

	log.Debug("basic authentication")
	var user db.User

	sql := h.db.Where("username = ?", auth.Username()).First(&user)
	if sql.Error != nil {
		if sql.Error == gorm.ErrRecordNotFound {
			response.New(w, r).Failed().Send(401)
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

	var pki db.PKI
	sql = h.db.Model(&db.PKI{}).Where("active = ? AND expires_at > ?", true, time.Now().UTC()).Order("created_at DESC").First(&pki)
	if sql.Error != nil {
		log.WithError(sql.Error).Error("unable to query database")
		w.WriteHeader(500)
		return
	}

	audience = r.URL.Query().Get("service")
	subject = r.URL.Query().Get("account")

	var newScopes = []docker.Scope{}

	scope := r.URL.Query().Get("scope")
	if scope != "" {
		scopes, _ = docker.ParseScope(scope)

		var user db.User
		var permissions []db.Permission

		// Grab the user that's auth
		sql := h.db.Model(&db.User{}).Preload(clause.Associations).Where("username = ?", subject).First(&user)
		if sql.Error != nil {
			log.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}

		// Bring all user ids and group ids into a single slice
		var entityIds []int64 = make([]int64, 0)
		entityIds = append(entityIds, user.ID)
		for _, e := range user.Groups {
			entityIds = append(entityIds, e.ID)
		}

		// Query for all permissions that match the entity ids
		sql = h.db.Model(&db.Permission{}).Where(h.db.Where("entity_id IN ?", entityIds))
		for _, scope := range scopes {
			if len(scope.Actions) == 1 && scope.Actions[0] == "pull" {
				scope.Actions = append(scope.Actions, "push")
			}

			namespaces := strings.Split(scope.Name, "/")
			namespaces = namespaces[0 : len(namespaces)-1]

			sql = sql.Where(h.db.Where("type = ?", "namespace").Where("name IN ?", namespaces).Where("action IN ?", scope.Actions).Or(
				h.db.Where("type = ?", scope.Type).Where("name = ?", scope.Name).Where("action IN ?", scope.Actions),
			))
		}

		sql = sql.Find(&permissions)
		if sql.Error != nil {
			log.WithError(sql.Error).Error("unable to query database")
			w.WriteHeader(500)
			return
		}

		for _, perm := range permissions {
			actions := []string{string(perm.Action)}

			if perm.Action == "push" {
				actions = append(actions, "pull")
			}

			if perm.Type == "namespace" {
				for _, scope := range scopes {
					if strings.HasPrefix(scope.Name, perm.Name) {
						newScopes = append(newScopes, docker.Scope{
							Type:    scope.Type,
							Name:    scope.Name,
							Actions: actions,
						})
					}
				}
			} else {
				newScopes = append(newScopes, docker.Scope{
					Type:    string(perm.Type),
					Name:    perm.Name,
					Actions: actions,
				})
			}
		}

		for _, s := range newScopes {
			log.WithFields(logrus.Fields{
				"type":    s.Type,
				"name":    s.Name,
				"actions": strings.Join(s.Actions, ","),
			}).Debug("reconciled permission")
		}
	}

	x5c := strings.ReplaceAll(pki.X509, "-----BEGIN CERTIFICATE-----", "")
	x5c = strings.ReplaceAll(x5c, "-----END CERTIFICATE-----", "")
	x5c = strings.ReplaceAll(x5c, "\n", "")

	signingMethod := ""
	if pki.Type == "ECDSA" {
		signingMethod = fmt.Sprintf("ES%d", pki.Bits)
	} else if pki.Type == "RSA" {
		signingMethod = fmt.Sprintf("RS%d", pki.Bits)
	} else {
		err := fmt.Errorf("invalid pki type: %s", pki.Type)
		log.WithError(err).Error("unknown signing algorithm")
		response.New(w, r).AddError(err).Send(500)
		return
	}

	log.Debugf("signing method: %s", signingMethod)

	t := jwt.New(jwt.GetSigningMethod(signingMethod))
	t.Claims = TokenClaims{
		Access: newScopes,
		StandardClaims: jwt.StandardClaims{
			Id:        uuid.NewString(),
			Audience:  audience,
			Issuer:    common.AppVersion.Name,
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: time.Now().UTC().Add(300 * time.Second).Unix(),
			NotBefore: time.Now().UTC().Unix(),
			Subject:   subject,
		},
	}
	t.Header["x5c"] = []string{x5c}

	var key crypto.PrivateKey

	if pki.Type == "RSA" {
		key, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(pki.Private))
		if err != nil {
			log.WithError(err).Error("unable to parse pki private key")
			w.WriteHeader(500)
			return
		}
	} else if pki.Type == "ECDSA" {
		key, err = jwt.ParseECPrivateKeyFromPEM([]byte(pki.Private))
		if err != nil {
			log.WithError(err).Error("unable to parse pki private key")
			w.WriteHeader(500)
			return
		}
	}

	token, err := t.SignedString(key)
	if err != nil {
		log.WithError(err).Error("unable to sign token")
		w.WriteHeader(500)
		return
	}

	log.Trace(token)

	res := TokenResponse{
		Token:       token,
		AccessToken: token,
		ExpiresIn:   300,
		IssuedAt:    time.Now().UTC(),
	}

	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(500)
		log.WithError(err).Error("unable to encode json")
		return
	}
}
