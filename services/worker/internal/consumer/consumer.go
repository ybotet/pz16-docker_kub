package consumer

import (
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/ybotet/pz12-REST_vs_GraphQL/services/worker/internal/idempotency"
	"github.com/ybotet/pz12-REST_vs_GraphQL/services/worker/internal/processor"
	"github.com/ybotet/pz12-REST_vs_GraphQL/shared/models"
	"github.com/ybotet/pz12-REST_vs_GraphQL/shared/rabbit"
)

const (
	maxAttempts = 3
	mainQueue   = "task_jobs"
	dlqQueue    = "task_jobs_dlq"
)

type JobConsumer struct {
	client         *rabbit.RabbitClient
	logger         *logrus.Logger
	processor      *processor.TaskProcessor
	idempotency    *idempotency.Store
}

func NewJobConsumer(rabbitURL string, logger *logrus.Logger) (*JobConsumer, error) {
	client, err := rabbit.NewRabbitClient(rabbitURL)
	if err != nil {
		return nil, err
	}

	// Declarar cola principal (durable)
	if err := client.DeclareQueue(mainQueue, true); err != nil {
		client.Close()
		return nil, err
	}

	// Declarar DLQ (durable)
	if err := client.DeclareQueue(dlqQueue, true); err != nil {
		client.Close()
		return nil, err
	}

	// Configurar prefetch = 1
	if err := client.Channel.Qos(1, 0, false); err != nil {
		client.Close()
		return nil, err
	}

	return &JobConsumer{
		client:      client,
		logger:      logger,
		processor:   processor.NewTaskProcessor(logger),
		idempotency: idempotency.NewStore(5 * time.Minute),
	}, nil
}

func (c *JobConsumer) Start() error {
	messages, err := c.client.Channel.Consume(
		mainQueue,
		"worker",
		false, // auto-ack false
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	c.logger.WithField("queue", mainQueue).Info("Job consumer started, waiting for messages...")

	for msg := range messages {
		c.processJob(msg)
	}

	return nil
}

func (c *JobConsumer) processJob(msg amqp.Delivery) {
	c.logger.WithField("body", string(msg.Body)).Info("Received job message")

	// Parsear el job
	var job models.TaskJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		c.logger.WithError(err).Error("Failed to parse job")
		msg.Nack(false, false) // no requeue, mensaje mal formado
		return
	}

	// Verificar idempotencia
	if c.idempotency.IsProcessed(job.MessageID) {
		c.logger.WithField("message_id", job.MessageID).Warn("Duplicate message detected, skipping")
		msg.Ack(false) // Confirmar para no procesar de nuevo
		return
	}

	// Procesar la tarea
	err := c.processor.Process(&job)

	if err == nil {
		// Éxito
		c.idempotency.MarkProcessed(job.MessageID)
		c.logger.WithFields(logrus.Fields{
			"task_id":    job.TaskID,
			"message_id": job.MessageID,
		}).Info("Job processed successfully")
		msg.Ack(false)
		return
	}

	// Error: manejar retry
	c.logger.WithError(err).WithFields(logrus.Fields{
		"task_id":    job.TaskID,
		"attempt":    job.Attempt,
		"message_id": job.MessageID,
	}).Warn("Job processing failed")

	job.Attempt++

	if job.Attempt <= maxAttempts {
		// Reintentar
		c.retryJob(&job)
	} else {
		// Enviar a DLQ
		c.sendToDLQ(&job, err)
	}

	// Siempre hacer ack del mensaje original
	msg.Ack(false)
}

func (c *JobConsumer) retryJob(job *models.TaskJob) {
	c.logger.WithFields(logrus.Fields{
		"task_id":    job.TaskID,
		"attempt":    job.Attempt,
		"max":        maxAttempts,
		"message_id": job.MessageID,
	}).Info("Retrying job")

	body, _ := json.Marshal(job)
	if err := c.client.PublishJSON(mainQueue, body); err != nil {
		c.logger.WithError(err).Error("Failed to retry job")
	}
}

func (c *JobConsumer) sendToDLQ(job *models.TaskJob, originalErr error) {
	c.logger.WithFields(logrus.Fields{
		"task_id":    job.TaskID,
		"attempts":   job.Attempt - 1,
		"message_id": job.MessageID,
		"error":      originalErr.Error(),
	}).Warn("Max attempts reached, sending to DLQ")

	body, _ := json.Marshal(job)
	if err := c.client.PublishJSON(dlqQueue, body); err != nil {
		c.logger.WithError(err).Error("Failed to send to DLQ")
	}
}

func (c *JobConsumer) Close() {
	c.client.Close()
}