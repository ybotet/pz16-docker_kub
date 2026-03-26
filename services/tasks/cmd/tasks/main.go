package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ybotet/pz6_csrf_xss/services/tasks/internal/clients"
	"github.com/ybotet/pz6_csrf_xss/services/tasks/internal/handlers"
	"github.com/ybotet/pz6_csrf_xss/services/tasks/internal/repository"

	internalMiddleware "github.com/ybotet/pz6_csrf_xss/services/tasks/internal/middleware"
	"github.com/ybotet/pz6_csrf_xss/shared/logger"
	"github.com/ybotet/pz6_csrf_xss/shared/middleware"
)

func main() {
    tasksPort := os.Getenv("TASKS_PORT")
    if tasksPort == "" {
        tasksPort = "8082"
    }

    authAddr := os.Getenv("AUTH_GRPC_ADDR")
    if authAddr == "" {
        authAddr = "localhost:50051"
    }

    log := logger.New(logger.Config{
        ServiceName: "tasks",
        Environment: "development",
        LogLevel:    "debug",
        JSONFormat:  true,
    })

    // Conectar a PostgreSQL
    db, err := repository.NewPostgresConnection()
    if err != nil {
        log.Fatalf("Error conectando a PostgreSQL: %v", err)
    }
    defer db.Close()

    // Crear repositorio
    taskRepo := repository.NewPostgresTaskRepository(db)

    // Router
    r := mux.NewRouter()

    // Middlewares GLOBALES (en orden correcto)
    r.Use(middleware.RequestID)
    r.Use(middleware.Logging(log))
    r.Use(internalMiddleware.SecurityHeadersMiddleware) // <-- NUEVO

    // Conectar a Auth service
    authClient, err := clients.NewAuthClient(authAddr)
    if err != nil {
        log.Fatalf("Error conectando a Auth service: %v", err)
    }
    defer authClient.Close()

    // Middleware de autenticación (existente)
    authMiddleware := internalMiddleware.NewAuthMiddleware(authClient.GetClient())
    taskHandler := handlers.NewTaskHandler(taskRepo)

    // Health check (público)
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")

    // ===== RUTAS PROTEGIDAS =====
    // GET /tasks - Solo autenticación (no requiere CSRF)
    r.HandleFunc("/v1/tasks", 
        authMiddleware.Authenticate(taskHandler.GetTasks)).Methods("GET")
    
    // POST /tasks - Autenticación + CSRF
    r.HandleFunc("/v1/tasks", 
        authMiddleware.Authenticate(internalMiddleware.CSRFMiddleware(taskHandler.CreateTask))).Methods("POST")

    r.HandleFunc("/v1/tasks/search/vulnerable", 
        authMiddleware.Authenticate(taskHandler.SearchTasksVulnerable)).Methods("GET")
    r.HandleFunc("/v1/tasks/search", 
        authMiddleware.Authenticate(taskHandler.SearchTasks)).Methods("GET")

    log.Printf("Servidor Tasks escuchando en puerto %s", tasksPort)
    log.Fatal(http.ListenAndServe(":"+tasksPort, r))
}