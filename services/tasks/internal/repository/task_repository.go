package repository

import (
	"database/sql"
	"fmt"

	"github.com/ybotet/pz6_csrf_xss/shared/models"
)

type TaskRepository interface {
    GetAll() ([]models.Task, error)
    GetByID(id string) (*models.Task, error)
    GetByUserID(userID string) ([]models.Task, error) // NUEVO
    Create(task *models.Task) error
    Update(task *models.Task) error
    Delete(id string, userID string) error // ¡CAMBIADO! ahora requiere userID
    SearchByTitle(title string) ([]models.Task, error)
    SearchByTitleVulnerable(title string) ([]models.Task, error) // Para demo
}

type PostgresTaskRepository struct {
    db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
    return &PostgresTaskRepository{db: db}
}

// ===== NUEVO: GetByUserID =====
func (r *PostgresTaskRepository) GetByUserID(userID string) ([]models.Task, error) {
    query := `SELECT id, title, description, done, created_at, user_id 
              FROM tasks WHERE user_id = $1 ORDER BY created_at DESC`
    
    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("error getting tasks by user: %w", err)
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UserID)
        if err != nil {
            return nil, fmt.Errorf("error scanning task: %w", err)
        }
        tasks = append(tasks, t)
    }
    return tasks, nil
}

// GetAll - Obtener todas las tareas (sin filtrar por usuario - SOLO PARA ADMIN)
func (r *PostgresTaskRepository) GetAll() ([]models.Task, error) {
    query := `SELECT id, title, description, done, created_at, user_id 
              FROM tasks ORDER BY created_at DESC`
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("error getting all tasks: %w", err)
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UserID)
        if err != nil {
            return nil, fmt.Errorf("error scanning task: %w", err)
        }
        tasks = append(tasks, t)
    }
    return tasks, nil
}

// GetByID - Obtener tarea por ID
func (r *PostgresTaskRepository) GetByID(id string) (*models.Task, error) {
    var t models.Task
    query := `SELECT id, title, description, done, created_at, user_id 
              FROM tasks WHERE id = $1`
    
    err := r.db.QueryRow(query, id).Scan(
        &t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UserID)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // No encontrada
        }
        return nil, fmt.Errorf("error getting task by ID: %w", err)
    }
    return &t, nil
}

// Create - Crear nueva tarea (AHORA con user_id)
func (r *PostgresTaskRepository) Create(task *models.Task) error {
    query := `INSERT INTO tasks (id, title, description, done, created_at, user_id) 
              VALUES ($1, $2, $3, $4, $5, $6)`
    
    result, err := r.db.Exec(query, 
        task.ID, 
        task.Title, 
        task.Description, 
        task.Done, 
        task.CreatedAt,
        task.UserID) // NUEVO campo
    
    if err != nil {
        return fmt.Errorf("error inserting task: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("no rows affected")
    }
    
    return nil
}

// Update - Actualizar tarea (verificando user_id)
func (r *PostgresTaskRepository) Update(task *models.Task) error {
    query := `UPDATE tasks 
              SET title = $1, description = $2, done = $3 
              WHERE id = $4 AND user_id = $5` // Verificar user_id
    
    result, err := r.db.Exec(query, 
        task.Title, 
        task.Description, 
        task.Done, 
        task.ID,
        task.UserID) // Verificar que sea el dueño
    
    if err != nil {
        return fmt.Errorf("error updating task: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("task not found or not authorized")
    }
    
    return nil
}

// Delete - Eliminar tarea (verificando user_id)
func (r *PostgresTaskRepository) Delete(id string, userID string) error {
    query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
    
    result, err := r.db.Exec(query, id, userID)
    if err != nil {
        return fmt.Errorf("error deleting task: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("task not found or not authorized")
    }
    
    return nil
}

// VERSIÓN VULNERABLE (para demostración SQL injection)
func (r *PostgresTaskRepository) SearchByTitleVulnerable(title string) ([]models.Task, error) {
    // ¡MAL! Concatenación de strings - VULNERABLE A SQL INJECTION
    query := fmt.Sprintf(
        "SELECT id, title, description, done, created_at, user_id FROM tasks WHERE title = '%s'", 
        title)
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("error in vulnerable search: %w", err)
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UserID)
        if err != nil {
            return nil, fmt.Errorf("error scanning task: %w", err)
        }
        tasks = append(tasks, t)
    }
    return tasks, nil
}

// VERSIÓN SEGURA (parametrizada)
func (r *PostgresTaskRepository) SearchByTitle(title string) ([]models.Task, error) {
    // ¡BIEN! Consulta parametrizada - SEGURA
    query := `SELECT id, title, description, done, created_at, user_id 
              FROM tasks WHERE title = $1`
    
    rows, err := r.db.Query(query, title)
    if err != nil {
        return nil, fmt.Errorf("error in safe search: %w", err)
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UserID)
        if err != nil {
            return nil, fmt.Errorf("error scanning task: %w", err)
        }
        tasks = append(tasks, t)
    }
    return tasks, nil
}