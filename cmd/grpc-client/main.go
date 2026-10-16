package main

import (
	"context"
	"log"
	"time"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/api/proto/deployment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	conn, err := grpc.NewClient(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := deploymentv1.NewDeploymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	/*response, err := client.CreateDeployment(
		ctx,
		&deploymentv1.CreateDeploymentRequest{
			ProjectId: "6c938013-3b89-4dd2-8744-8e05714e8c52",
			Branch:    "main",
		},
	)*/
	response, err := client.GetDeployment(
		ctx,
		&deploymentv1.GetDeploymentRequest{
			DeploymentId: "fdbf3ca3-a1de-4613-a5a4-22b5ffcb0c67",
		},
	)

	if err != nil {
		log.Printf("code=%s message=%s", status.Code(err), status.Convert(err).Message())
		return
	}

	log.Printf("response: %v", response)
}
