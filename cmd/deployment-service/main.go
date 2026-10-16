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
		panic(err)
	}
}

func run() error {
	cfg, err := config.Load()
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
	handler := grpcserver.NewServer(deploymentService)

	listener, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Print(err)
		return err
	}

	server := grpc.NewServer()
	deploymentv1.RegisterDeploymentServiceServer(server, handler)

	log.Println("gRPC listening on 127.0.0.1:50051")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve(listener)
	}()

	var serveErr error

	select {
	case <-ctx.Done():
		stop()
		log.Println("shutdown signal received")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr = fmt.Errorf("grpc serve: %w", err)
		}
	}

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

	return serveErr
}
