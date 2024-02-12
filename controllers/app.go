package controllers

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/goidp/models"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

// Build vars populated with lflags
var (
	AppBuild     = "0"
	AppName      = "idp"
	ChartVersion = "0.0.0"
	AppVersion   = "0.0.0.0"
)

type App struct {
	server *http.Server
	router *mux.Router
	Events interface {
		GetEvents(pageNumber int, pageSize int) []*models.Event
		GetEventsCount(s models.EventSeverity) int
		Create(e *models.Event) error
		CreateUnsuccessfulLoginEvent(username, domain, ip string) error
		CreateSuccessfulLoginEvent(username, domain, ip string) error
		CreateUserEvent(method, username, domain string) error
		CreateJWTEvent(username, domain string) error
	}
	Users interface {
		Create(u *models.User) error
		ValidateUsername(u string) error
		UpdateUserByNameOrID(nameOrID, username, password string, roles []string) (*models.User, error)
		UpdateUser(u *models.User) error
		GetUserByNameOrID(nameOrID string) (*models.User, error)
		GetUsers() ([]*models.User, error)
		GetUsersID() []uint
		DeleteUser(user *models.User) (err error)
		DeleteUserByID(id int) error
		DeleteUserByName(name string) error
		DeleteUserByNameOrID(nameOrID string) error
		DeleteAllUsers() error
		GetAndValidateUser(username string, password string) (*models.User, bool)
		AddDefaultUser()
	}
	Roles interface {
		AddDefaultRoles()
	}
	extUsers map[string]models.RoleList
	config   *Config
}

type Config struct {
	LogLevel              string
	WriteTimeout          int
	ReadTimeout           int
	IdleTimeout           int
	Host                  string
	Port                  string
	Secret                string
	SignKey               *rsa.PrivateKey
	VerifyKey             *rsa.PublicKey
	AccessTokenExpireTime time.Duration
	RenewTokenExpireTime  time.Duration
	TrustedPublicKeys     []*rsa.PublicKey
}

func (a *App) setRouters() {
	a.router.Use(a.loggingMiddleware)
	a.router.Use(a.jsonapiMiddleware)
	a.router.HandleFunc("/versions", a.GetVersions).Methods(http.MethodGet)
	baseURL := fmt.Sprintf("/%s", ApiVersion)
	base := a.router.PathPrefix(baseURL).Subrouter()
	base.HandleFunc("/session", a.SessionHandler).Methods(http.MethodPost, http.MethodDelete)
	base.HandleFunc("/renew", a.RenewTokenHandler).Methods(http.MethodPost)
	usersRouter := base.PathPrefix("/user").Subrouter()

	usersRouter.Use(func(next http.Handler) http.Handler {
		return a.jwtMiddleware(next)
	})

	usersRouter.HandleFunc("", a.UsersHandler).Methods(http.MethodGet, http.MethodPost)
	usersRouter.HandleFunc("/{id}", a.UserHandler).Methods(http.MethodDelete, http.MethodPatch, http.MethodGet)

	eventRouter := base.PathPrefix("/event").Subrouter()
	eventRouter.Use(func(next http.Handler) http.Handler {
		return a.jwtMiddleware(next)
	})
	eventRouter.HandleFunc("", a.EventsHandler).Methods(http.MethodGet)

	systemRouter := base.PathPrefix("/system").Subrouter()
	systemRouter.Use(func(next http.Handler) http.Handler {
		return a.jwtMiddleware(next)
	})
	systemRouter.HandleFunc("", a.SystemHandler).Methods(http.MethodGet)
}

func NewApp(db *gorm.DB, c *Config) *App {
	var a App
	var level log.Level

	a.config = c

	// parse log level, default to warning
	level, err := log.ParseLevel(strings.ToLower(a.config.LogLevel))
	if err != nil {
		level = log.WarnLevel
	}

	log.SetLevel(level)
	log.SetFormatter(&log.JSONFormatter{
		PrettyPrint: true,
	})

	a.router = mux.NewRouter().StrictSlash(true)

	a.setRouters()

	a.server = &http.Server{
		Handler:      a.router,
		Addr:         fmt.Sprintf("%s:%s", a.config.Host, a.config.Port),
		WriteTimeout: time.Duration(a.config.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(a.config.ReadTimeout) * time.Second,
		IdleTimeout:  time.Duration(a.config.IdleTimeout) * time.Second,
	}
	a.Roles = &models.RoleRepo{DB: db}
	a.Users = &models.UserRepo{DB: db}
	a.Events = &models.EventRepo{DB: db}
	a.extUsers = make(map[string]models.RoleList)
	return &a
}

func (a *App) Run() {
	log.WithFields(log.Fields{
		"app_name":      AppName,
		"app_version":   AppVersion,
		"chart_version": ChartVersion,
		"api_version":   ApiVersion,
		"build_id":      AppBuild,
		"host":          a.server.Addr,
	}).Info("Listening...")

	// Handle server shutdown gracefully
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		err := a.server.ListenAndServe()
		if err != nil {
			log.WithFields(log.Fields{
				"host":  a.server.Addr,
				"error": err,
			}).Fatal("shutting down...")
		}
	}()

	// the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m
	var wait = 15 * time.Second

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = a.server.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
}

func (a *App) AddDefaultUserAndRoles() {
	a.Roles.AddDefaultRoles()
	a.Users.AddDefaultUser()
}

// loggingMiddleware is a middleware function that logs requests to the API
func (a *App) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.WithFields(log.Fields{
			"requestUri": r.RequestURI,
			"from":       r.Host,
		}).Info("Got API call")
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
