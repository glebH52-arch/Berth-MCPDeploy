package grpc

import (
	"context"
	"errors"
	"log"
	"strings"
	"uuid"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/api/proto/deployment/v1"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	deploymentv1.UnimplementedDeploymentServiceServer
	deploymentService *service.DeploymentService
}

var _ deploymentv1.DeploymentServiceServer = (*Server)(nil)

func NewServer(deploymentService *service.DeploymentService) *Server {
	return &Server{
		deploymentService: deploymentService,
	}
}

func statusToProto(s domain.DeploymentStatus) deploymentv1.DeploymentStatus {
	switch s {
	case domain.DeploymentStatusQueued:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_QUEUED
	case domain.DeploymentStatusCloning:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_CLONING
	case domain.DeploymentStatusBuilding:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_BUILDING
	case domain.DeploymentStatusStarting:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_STARTING
	case domain.DeploymentStatusHealthChecking:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_HEALTHCHECKING
	case domain.DeploymentStatusRunning:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING
	case domain.DeploymentStatusFailed:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_FAILED
	default:
		return deploymentv1.DeploymentStatus_DEPLOYMENT_STATUS_UNSPECIFIED
	}
}

func (s *Server) GetDeployment(ctx context.Context, req *deploymentv1.GetDeploymentRequest) (*deploymentv1.GetDeploymentResponse, error) {

	reqID := req.GetDeploymentId()
	id, err := uuid.Parse(reqID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "deployment_id must be a valid UUID")
	}
	deployment, err := s.deploymentService.GetDeployment(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrDeploymentNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else {
			return nil, status.Error(codes.Internal, "failed to get deployment")
		}
	}
	result := &deploymentv1.Deployment{
		DeploymentId:  deployment.ID.String(),
		ProjectId:     deployment.ProjectID.String(),
		RepositoryUrl: deployment.RepositoryURL,
		Branch:        deployment.Branch,
		CommitSha:     deployment.CommitSHA,
		Url:           deployment.URL,
		CreatedAt:     timestamppb.New(deployment.CreatedAt),
		UpdatedAt:     timestamppb.New(deployment.UpdatedAt),
	}
	if deployment.RunningAt != nil {
		result.RunningAt = timestamppb.New(*deployment.RunningAt)
	}
	if deployment.FailedAt != nil {
		result.Failure = &deploymentv1.DeploymentFailure{
			FailedAt:     timestamppb.New(*deployment.FailedAt),
			FailedStage:  deployment.FailureStage,
			ErrorCode:    deployment.ErrorClass,
			ErrorMessage: deployment.ErrorMessage,
			Diagnostic:   deployment.LogExcerpt,
		}
	}
	result.Status = statusToProto(deployment.Status)

	return &deploymentv1.GetDeploymentResponse{Deployment: result}, nil
}

func (s *Server) CreateDeployment(ctx context.Context, req *deploymentv1.CreateDeploymentRequest) (*deploymentv1.CreateDeploymentResponse, error) {
	reqID := req.GetProjectId()
	projectID, err := uuid.Parse(reqID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "project_id must be a valid UUID")
	}
	branch := strings.TrimSpace(req.GetBranch())
	if branch == "" {
		return nil, status.Error(codes.InvalidArgument, "branch must not be empty")
	}
	deployment, err := s.deploymentService.CreateDeployment(ctx, projectID, branch)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProjectNotFound):
			return nil, status.Error(codes.NotFound, "project not found")
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, "request canceled")
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "request deadline exceeded")
		default:
			log.Printf("create deployment failed: %v", err)
			return nil, status.Error(codes.Internal, "failed to create deployment")
		}
	}
	return &deploymentv1.CreateDeploymentResponse{DeploymentId: deployment.ID.String(), Status: statusToProto(deployment.Status)}, nil
}
