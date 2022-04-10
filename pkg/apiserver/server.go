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
	// Basic Auth
	api.Path("/token").Methods("GET").HandlerFunc(handlers.Token)
	// Bearer / OAuth2 Auth
	api.Path("/token").Methods("POST").HandlerFunc(handlers.BearerToken)

	api.Path("/groups").Methods("POST").HandlerFunc(handlers.Root)
	api.Path("/groups/:id").Methods("DELETE").HandlerFunc(handlers.Root)

	api.Path("/grant").Methods("POST").HandlerFunc(handlers.Grant)
	api.Path("/revoke").Methods("POST").HandlerFunc(handlers.Grant)

	api.Path("/admin/user/{user}/{type}/{name}/{action}").Methods("POST").HandlerFunc(handlers.Grant)
	api.Path("/admin/group/{group}/{type}/{name}/{action}").Methods("DELETE").HandlerFunc(handlers.Revoke)

	// Below this point is where the server is started and graceful shutdown occurs.

	router.NotFoundHandler = router.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: ghandlers.CORS()(router),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("listen: %s\n", err)
		}
	}()
	a.log.WithField("port", a.port).Info("Starting API Server")

	<-a.ctx.Done()

	a.log.Info("Shutting down API Server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		a.log.WithError(err).Error("Unable to shutdown the API server gracefully")
		return err
	}

	return nil
}
