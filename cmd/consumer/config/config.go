package config

import (
	"fmt"
	"os"

	"github.com/Julia-Marcal/eventcounter/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQConnString string
	QueueName          string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		logger.Warning("Arquivo .env não encontrado, lendo variáveis de ambiente")
	}

	rabbitmq_user := getEnv("RABBITMQ_USER", "guest")
	rabbitmq_password := getEnv("RABBITMQ_PASSWORD", "guest")
	rabbitmq_host := getEnv("RABBITMQ_HOST", "localhost")
	rabbitmq_port := getEnv("RABBITMQ_PORT", "5672")
	queue_name := getEnv("RABBITMQ_QUEUE_NAME", "eventcountertest")

	cfg := &Config{
		RabbitMQConnString: fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmq_user, rabbitmq_password, rabbitmq_host, rabbitmq_port),
		QueueName:          queue_name,
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
