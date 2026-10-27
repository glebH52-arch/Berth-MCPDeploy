package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/config"
	"github.com/glebH52-arch/agent-native-paas/internal/worker"
	kafkamessaging "github.com/glebH52-arch/agent-native-paas/internal/worker/messaging/kafka"
)

func main() {
	if err := run(); err != nil {
		log.Panic(err)
	}
}

func run() error {
	cfg, err := config.LoadWorker()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	reader := kafkamessaging.NewReader(cfg.KafkaBrokers)
	defer reader.Close()

	writer := kafkamessaging.NewWriter(cfg.KafkaBrokers)
	defer writer.Close()

	publisher := kafkamessaging.NewKafkaEventPublisher(writer)
	handler := worker.NewWorker(publisher)
	consumer := kafkamessaging.NewKafkaEventConsumer(reader, handler)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	return consumer.Run(ctx)
}
