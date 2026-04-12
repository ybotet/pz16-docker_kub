package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/ybotet/pz12-REST_vs_GraphQL/services/worker/internal/consumer"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	rabbitURL := getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")

	logger.WithField("rabbit_url", rabbitURL).Info("Starting Job Worker")

	consumer, err := consumer.NewJobConsumer(rabbitURL, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create consumer")
	}
	defer consumer.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := consumer.Start(); err != nil {
			logger.WithError(err).Fatal("Consumer error")
		}
	}()

	logger.Info("Job worker is running. Press Ctrl+C to stop.")
	<-sigChan
	logger.Info("Shutting down job worker...")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}