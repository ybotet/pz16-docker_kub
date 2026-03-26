package models

import "time"

type Task struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Done        bool      `json:"done"`
    CreatedAt   time.Time `json:"created_at"`
    UserID      string    `json:"user_id"` // NUEVO: para asociar tarea al usuario
}