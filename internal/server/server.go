package server

import (
	"context"
	"net"
	"net/http"

	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	"github.com/EduardMikhrin/university-booking-project/internal/data"
	httpSwagger "github.com/swaggo/http-swagger"
	"gitlab.com/distributed_lab/logan/v3"
)

type Server struct {
	log       *logan.Entry
	db        data.MasterQ
	cache     cache.CacheQ
	listener  net.Listener
	jwtConfig JWT
	router    *http.ServeMux
}

func NewServer(log *logan.Entry, db data.MasterQ, cache cache.CacheQ, listener net.Listener, jwtConfig JWT) *Server {
	s := &Server{
		log:       log,
		db:        db,
		cache:     cache,
		listener:  listener,
		jwtConfig: jwtConfig,
		router:    http.NewServeMux(),
	}
	s.mountRoutes()
	return s
}

func (s *Server) mountRoutes() {
	// API v1 base path
	apiV1 := http.NewServeMux()

	// Authentication routes (public - no middleware)
	apiV1.HandleFunc("POST /auth/login", s.handleLogin)
	apiV1.HandleFunc("POST /auth/register", s.handleRegister)

	// Authentication routes (require authentication)
	apiV1.HandleFunc("GET /auth/me", s.userMiddleware(s.handleGetMe))
	apiV1.HandleFunc("POST /auth/logout", s.userMiddleware(s.handleLogout))

	// Reservation routes (require authentication)
	apiV1.HandleFunc("GET /reservations", s.userMiddleware(s.handleGetReservations))
	apiV1.HandleFunc("GET /reservations/{id}", s.userMiddleware(s.handleGetReservation))
	apiV1.HandleFunc("GET /reservations/user/{userId}", s.userMiddleware(s.handleGetUserReservations))
	apiV1.HandleFunc("POST /reservations", s.userMiddleware(s.handleCreateReservation))
	apiV1.HandleFunc("PATCH /reservations/{id}", s.userMiddleware(s.handleUpdateReservation))
	apiV1.HandleFunc("PATCH /reservations/{id}/status", s.userMiddleware(s.handleUpdateReservationStatus))
	apiV1.HandleFunc("DELETE /reservations/{id}", s.userMiddleware(s.handleDeleteReservation))

	// Table routes (require authentication)
	apiV1.HandleFunc("GET /tables", s.userMiddleware(s.handleGetTables))
	apiV1.HandleFunc("GET /tables/{id}", s.userMiddleware(s.handleGetTable))
	apiV1.HandleFunc("GET /tables/available", s.userMiddleware(s.handleGetAvailableTables))
	apiV1.HandleFunc("PATCH /tables/{id}/availability", s.userMiddleware(s.handleUpdateTableAvailability))

	// Report routes (Admin only)
	apiV1.HandleFunc("GET /reports/monthly", s.adminMiddleware(s.handleGetMonthlyReports))
	apiV1.HandleFunc("GET /reports/monthly/{month}", s.adminMiddleware(s.handleGetMonthlyReport))

	// User routes (require authentication)
	apiV1.HandleFunc("GET /users/{id}", s.userMiddleware(s.handleGetUser))
	apiV1.HandleFunc("PATCH /users/{id}", s.userMiddleware(s.handleUpdateUser))

	// Mount API v1 under /api/v1
	s.router.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1))
	s.router.Handle("/swagger/", http.StripPrefix("/swagger/", httpSwagger.WrapHandler))
}

// Run starts the HTTP server and blocks until an error occurs
func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Handler: corsMiddleware(s.router),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	s.log.WithField("address", s.listener.Addr().String()).Info("starting server")
	return server.Serve(s.listener)
}
