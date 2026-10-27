package config

import (
	"fmt"
	"os"
	"strings"
)

type DeploymentConfig struct {
	DatabaseURL  string
	GRPCAddr     string
	KafkaBrokers []string
}

type WorkerConfig struct {
	KafkaBrokers []string
}

type PublisherConfig struct {
	DatabaseURL  string
	KafkaBrokers []string
}

func LoadDeployment() (*DeploymentConfig, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	grpcAddr := strings.TrimSpace(os.Getenv("GRPC_ADDR"))
	if grpcAddr == "" {
		return nil, fmt.Errorf("GRPC_ADDR is required")
	}

	brokers, err := loadKafkaBrokers()
	if err != nil {
		return nil, err
	}

	return &DeploymentConfig{
		DatabaseURL:  databaseURL,
		GRPCAddr:     grpcAddr,
		KafkaBrokers: brokers,
	}, nil
}

func LoadPublisher() (*PublisherConfig, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	brokers, err := loadKafkaBrokers()
	if err != nil {
		return nil, err
	}

	return &PublisherConfig{
		DatabaseURL:  databaseURL,
		KafkaBrokers: brokers,
	}, nil
}

func LoadWorker() (*WorkerConfig, error) {
	brokers, err := loadKafkaBrokers()
	if err != nil {
		return nil, err
	}

	return &WorkerConfig{
		KafkaBrokers: brokers,
	}, nil
}

func loadKafkaBrokers() ([]string, error) {
	raw := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if raw == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}

	brokers := strings.Split(raw, ",")

	for i, broker := range brokers {
		brokers[i] = strings.TrimSpace(broker)

		if brokers[i] == "" {
			return nil, fmt.Errorf(
				"empty Kafka broker at index %d",
				i,
			)
		}
	}

	return brokers, nil
}
