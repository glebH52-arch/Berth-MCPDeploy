package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/config"
	gitresolver "github.com/glebH52-arch/agent-native-paas/internal/deployment/git"
	kafkamessaging "github.com/glebH52-arch/agent-native-paas/internal/deployment/messaging/kafka"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository/pgsql"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/service"

	"log"
	"net"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/api/proto/deployment/v1"

	grpcserver "github.com/glebH52-arch/agent-native-paas/internal/deployment/transport/grpc"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Panic(err)
	}
}

func run() error {
	cfg, err := config.LoadDeployment()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := pgsql.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	projectPgsqlRepo := pgsql.NewPostgresProjectRepository(pool)
	//projectService := service.NewProjectService(projectPgsqlRepo)
	deploymentRepo := pgsql.NewPostgresDeploymentRepository(pool)
	resolver := &gitresolver.Resolver{}
	deploymentService := service.NewDeploymentService(
		deploymentRepo,
		projectPgsqlRepo,
		resolver,
	)
	reader := kafkamessaging.NewReader(cfg.KafkaBrokers)
	defer reader.Close()

	consumer := kafkamessaging.NewKafkaEventConsumer(reader, deploymentService)

	kafkaError := make(chan error, 1)

	handler := grpcserver.NewServer(deploymentService)

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Print(err)
		return err
	}

	server := grpc.NewServer()
	deploymentv1.RegisterDeploymentServiceServer(server, handler)

	log.Printf("gRPC listening on %s", cfg.GRPCAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.Serve(listener)
	}()

	var serveErr error
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		kafkaError <- consumer.Run(ctx)
	}()
	select {
	case <-ctx.Done():

		log.Println("shutdown signal received")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr = fmt.Errorf("grpc serve: %w", err)
		}
	case err := <-kafkaError:
		if err != nil {
			serveErr = fmt.Errorf("deployment event consumer: %w", err)
		}
	}
	stop()
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-done:
		log.Println("gRPC stopped gracefully")

	case <-timer.C:
		log.Println("shutdown timeout: forcing gRPC stop")
		server.Stop()
		<-done
	}
	<-consumerDone
	return serveErr
}
