package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ekristen/dockit/pkg/apiserver/handlers"
	"github.com/ekristen/dockit/pkg/apiserver/middleware"
)

type apiServer struct {
	ctx  context.Context
	log  *logrus.Entry
	db   *gorm.DB
	port int
}

func Register(ctx context.Context, log *logrus.Entry, db *gorm.DB, port int) *apiServer {
	return &apiServer{
		ctx:  ctx,
		log:  log,
		db:   db,
		port: port,
	}
}

func (a *apiServer) Start() error {
	handlers := handlers.New(a.db)
	defaultm := middleware.NewToken(a.log)

	router := mux.NewRouter().StrictSlash(true)

	router.Use(defaultm.RequestID)
	router.Use(middleware.LoggingMiddleware2(a.log))

	router.Path("/").HandlerFunc(handlers.Root)

	api := router.PathPrefix("/v2").Subrouter()

	// Token with Basic Auth
	api.Path("/token").Methods("GET").HandlerFunc(handlers.Token)

	// Token with Bearer/OAuth2 Auth
	api.Path("/token").Methods("POST").HandlerFunc(handlers.BearerToken)

	// Grant / Revoke Permissions
	api.Path("/admin/{rbac_type}:{rbac_entity}/{type}:{name}:{action}").Methods("PUT").HandlerFunc(handlers.Permission)
	api.Path("/admin/{rbac_type}:{rbac_entity}/{type}:{name}:{action}").Methods("DELETE").HandlerFunc(handlers.Permission)

	// Create User / Group
	api.Path("/admin/{rbac_type}:{rbac_entity}").Methods("PUT").HandlerFunc(handlers.Root)

	// Change Password / Enable / Disable
	api.Path("/admin/{rbac_type}:{rbac_entity}/{action}").Methods("PUT").HandlerFunc(handlers.Action)

	api.Path("/admin/{rbac_type}:{rbac_entity}/{rbac_type_2}:{rbac_entity_2}/{action}").Methods("PUT").HandlerFunc(handlers.Action)

	// PKI Cert Bundle
	api.Path("/certs/{format}").Methods("GET").HandlerFunc(handlers.PKICerts)

	// Note: this allows not found urls to be logged via the middleware
	// It **HAS** to be defined after all other paths are defined.
	router.NotFoundHandler = router.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	// Below this point is where the server is started and graceful shutdown occurs.

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: ghandlers.CORS()(router),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("listen: %s\n", err)
		}
	}()
	a.log.WithField("port", a.port).Info("starting api server")

	<-a.ctx.Done()

	a.log.Info("shutting down the api server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		a.log.WithError(err).Error("unable to shutdown the api server gracefully")
		return err
	}

	return nil
}
