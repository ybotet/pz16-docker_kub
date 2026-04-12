package models

import "time"

// TaskJob representa una tarea pesada en la cola de trabajos
type TaskJob struct {
	Job       string    `json:"job"`        // "process_task"
	TaskID    string    `json:"task_id"`    // ID de la tarea a procesar
	Attempt   int       `json:"attempt"`    // Número de intento (1, 2, 3...)
	MessageID string    `json:"message_id"` // UUID único para idempotencia
	CreatedAt time.Time `json:"created_at"` // Cuando se creó el job
}

// JobResult representa el resultado del procesamiento
type JobResult struct {
	Success     bool   `json:"success"`
	MessageID   string `json:"message_id"`
	TaskID      string `json:"task_id"`
	Attempt     int    `json:"attempt"`
	Error       string `json:"error,omitempty"`
	ProcessedAt string `json:"processed_at"`
}