package handlers

import (
	"encoding/json"
	"html" // NUEVO: para escapar HTML
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ybotet/pz6_csrf_xss/services/tasks/internal/middleware" // Para GetUserID
	"github.com/ybotet/pz6_csrf_xss/services/tasks/internal/repository"
	"github.com/ybotet/pz6_csrf_xss/shared/models"
)

type TaskHandler struct {
    repo repository.TaskRepository
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
    return &TaskHandler{repo: repo}
}

// GetTasks - Obtener todas las tareas del usuario autenticado
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
    // Obtener user ID del contexto (del token JWT)
    userID := middleware.GetUserID(r.Context())
    
    // Usar el método que filtra por usuario
    tasks, err := h.repo.GetByUserID(userID)
    if err != nil {
        log.Printf("Error getting tasks: %v", err) // Para debugging
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}

// CreateTask - Crear nueva tarea (con sanitización XSS)
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    var task models.Task
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // ===== SANITIZACIÓN XSS =====
    task.Title = sanitizeInput(task.Title)
    task.Description = sanitizeInput(task.Description)
    
    // Obtener user ID del contexto
    userID := middleware.GetUserID(r.Context())
    
    task.ID = uuid.New().String()
    task.CreatedAt = time.Now()
    task.Done = false
    task.UserID = userID // Asignar el usuario autenticado
    
    if err := h.repo.Create(&task); err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

// SearchTasksVulnerable - ENDPOINT VULNERABLE (solo para demostración SQL injection)
func (h *TaskHandler) SearchTasksVulnerable(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Query().Get("title")
    if title == "" {
        http.Error(w, "title parameter required", http.StatusBadRequest)
        return
    }
    
    // Usar versión vulnerable (solo para demo)
    repo, ok := h.repo.(*repository.PostgresTaskRepository)
    if !ok {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    tasks, err := repo.SearchByTitleVulnerable(title)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(tasks)
}

// SearchTasks - ENDPOINT SEGURO (parametrizado)
func (h *TaskHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Query().Get("title")
    if title == "" {
        http.Error(w, "title parameter required", http.StatusBadRequest)
        return
    }
    
    tasks, err := h.repo.SearchByTitle(title)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(tasks)
}

// sanitizeInput escapa caracteres HTML para prevenir XSS
func sanitizeInput(input string) string {
    return html.EscapeString(input)
}