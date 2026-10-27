package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/config"
	kafkamessaging "github.com/glebH52-arch/agent-native-paas/internal/deployment/messaging/kafka"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/outbox"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository/pgsql"
)

func main() {
	if err := run(); err != nil {
		log.Panic(err)
	}
}

func run() error {
	cfg, err := config.LoadPublisher()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := pgsql.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	outboxRepo := pgsql.NewPostgresOutboxRepository(pool)
	writer := kafkamessaging.NewWriter(cfg.KafkaBrokers)
	defer writer.Close()
	eventPublisher := kafkamessaging.NewKafkaEventPublisher(writer)
	publisherID := uuid.New().String()
	pub := outbox.NewPublisher(outboxRepo, eventPublisher, publisherID)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	return pub.Run(ctx)
}
