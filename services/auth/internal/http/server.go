package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/ybotet/pz6_csrf_xss/services/auth/internal/auth"
	"github.com/ybotet/pz6_csrf_xss/services/auth/internal/http/handlers"
	"github.com/ybotet/pz6_csrf_xss/shared/middleware" // Tus middlewares
)

type Server struct {
    server  *http.Server
    logger  *logrus.Logger
    handler *handlers.AuthHandler
}

func NewServer(authService *auth.Service, logger *logrus.Logger, port string) *Server {
    // Crear handler
    authHandler := handlers.NewAuthHandler(authService, logger)
    
    // Router con tus middlewares existentes
    r := mux.NewRouter()
    
    // Usar tus middlewares en el orden correcto
    r.Use(middleware.RequestID)      // Primero, generar request ID
    r.Use(middleware.Logging(logger)) // Luego, loguear acceso (¡NOTA: es Logging, no AccessLog!)
    
    // Rutas
    r.HandleFunc("/v1/auth/login", authHandler.Login).Methods("POST")
    r.HandleFunc("/health", authHandler.Health).Methods("GET")
    
    // Servidor HTTP
    httpServer := &http.Server{
        Addr:         ":" + port,
        Handler:      r,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }
    
    return &Server{
        server:  httpServer,
        logger:  logger,
        handler: authHandler,
    }
}

func (s *Server) Start() error {
    s.logger.WithField("port", s.server.Addr).Info("📡 Servidor HTTP auth iniciado")
    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("Apagando servidor HTTP...")
    return s.server.Shutdown(ctx)
}