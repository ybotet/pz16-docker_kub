package processor

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ybotet/pz12-REST_vs_GraphQL/shared/models"
)

type TaskProcessor struct {
	logger *logrus.Logger
}

func NewTaskProcessor(logger *logrus.Logger) *TaskProcessor {
	return &TaskProcessor{
		logger: logger,
	}
}

// Process ejecuta la tarea pesada. Retorna error si falla.
func (p *TaskProcessor) Process(job *models.TaskJob) error {
	p.logger.WithFields(logrus.Fields{
		"task_id":    job.TaskID,
		"message_id": job.MessageID,
		"attempt":    job.Attempt,
	}).Info("Starting heavy task processing")

	// Simular trabajo pesado (2-5 segundos)
	workTime := time.Duration(2+rand.Intn(3)) * time.Second
	time.Sleep(workTime)

	
	// MODO PRUEBA DLQ: task_id = "dlq_test" 
	// SIEMPRE falla para probar Dead Letter Queue
	
	if job.TaskID == "dlq_test" {
		return fmt.Errorf("DLQ TEST: simulated permanent error for task_id=%s", job.TaskID)
	}

	
	// MODO PRUEBA: task_id contiene "fail"
	// SIEMPRE falla (error permanente)
	
	if contains(job.TaskID, "fail") {
		return fmt.Errorf("simulated permanent error for task_id=%s", job.TaskID)
	}

	
	// MODO NORMAL: error aleatorio 20%
	
	if rand.Float64() < 0.2 {
		return fmt.Errorf("simulated random error (20%% chance)")
	}

	p.logger.WithFields(logrus.Fields{
		"task_id":    job.TaskID,
		"message_id": job.MessageID,
		"duration":   workTime,
	}).Info("Heavy task completed successfully")

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && 
		(len(substr) == 0 || (len(s) >= len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr) && contains(s[1:], substr)))))
}