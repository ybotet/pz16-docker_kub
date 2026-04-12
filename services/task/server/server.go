// services/task/server/server.go
package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/ybotet/pz12-REST_vs_GraphQL/services/task/handlers"
	"github.com/ybotet/pz12-REST_vs_GraphQL/shared/rabbit"
	"github.com/ybotet/pz12-REST_vs_GraphQL/shared/repository"
)

type RESTServer struct {
	port         string
	handler      *handlers.TaskHandler
	logger       *logrus.Logger
	rabbitURL    string
	queueName    string
	rabbitClient *rabbit.RabbitClient
	repo         *repository.PostgresTaskRepository // Guardar referencia
}

func NewRESTServer(
	port string,
	repo *repository.PostgresTaskRepository,
	logger *logrus.Logger,
	rabbitURL string,
	queueName string,
) *RESTServer {
	return &RESTServer{
		port:      port,
		handler:   nil,
		logger:    logger,
		rabbitURL: rabbitURL,
		queueName: queueName,
		repo:      repo, // Guardar repo
	}
}

func (s *RESTServer) InitRabbitMQ() error {
	s.logger.Info("Connecting to RabbitMQ...")
	
	client, err := rabbit.NewRabbitClient(s.rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	
	// Declarar la cola (durable)
	if err := client.DeclareQueue(s.queueName, true); err != nil {
		client.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	s.rabbitClient = client
	s.logger.WithField("queue", s.queueName).Info("RabbitMQ connected and queue declared")
	
	// Crear el handler CON el cliente RabbitMQ
	s.handler = handlers.NewTaskHandler(s.repo, s.logger, s.rabbitClient, s.queueName)
	
	return nil
}

func (s *RESTServer) setupRoutes() *mux.Router {
	// Ya no creamos handler aquí, ya está creado en InitRabbitMQ
	r := mux.NewRouter()
	api := r.PathPrefix("/v1").Subrouter()
	api.HandleFunc("/tasks", s.handler.ListTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", s.handler.GetTask).Methods("GET")
	api.HandleFunc("/tasks", s.handler.CreateTask).Methods("POST")
	api.HandleFunc("/tasks/{id}", s.handler.UpdateTask).Methods("PATCH")
	api.HandleFunc("/tasks/{id}", s.handler.DeleteTask).Methods("DELETE")
	api.HandleFunc("/jobs/process-task", s.handler.CreateJob).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	return r
}

func (s *RESTServer) Start(repo *repository.PostgresTaskRepository) error {
	// Intentar conectar a RabbitMQ
	if err := s.InitRabbitMQ(); err != nil {
		s.logger.WithError(err).Warn("Failed to initialize RabbitMQ, events will not be published")
		// Si falla RabbitMQ, crear handler sin cliente
		s.handler = handlers.NewTaskHandler(s.repo, s.logger, nil, s.queueName)
	}
	
	router := s.setupRoutes()
	addr := fmt.Sprintf(":%s", s.port)
	s.logger.WithField("port", s.port).Info("Starting REST server on http://localhost" + addr)
	return http.ListenAndServe(addr, router)
}