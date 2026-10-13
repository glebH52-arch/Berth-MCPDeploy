package domain

type DeploymentStatus string

const (
	DeploymentStatusQueued            DeploymentStatus = "queued"
	DeploymentStatusCloning           DeploymentStatus = "cloning"
	DeploymentStatusCloneFailed       DeploymentStatus = "clone_failed"
	DeploymentStatusBuilding          DeploymentStatus = "building"
	DeploymentStatusBuildFailed       DeploymentStatus = "build_failed"
	DeploymentStatusStarting          DeploymentStatus = "starting"
	DeploymentStatusStartFailed       DeploymentStatus = "start_failed"
	DeploymentStatusHealthChecking    DeploymentStatus = "healthchecking"
	DeploymentStatusHealthCheckFailed DeploymentStatus = "healthcheck_failed"
	DeploymentStatusRunning           DeploymentStatus = "running"
)

var allowedTransitions = map[DeploymentStatus][]DeploymentStatus{
	DeploymentStatusQueued: {
		DeploymentStatusCloning,
	},
	DeploymentStatusCloning: {
		DeploymentStatusBuilding,
		DeploymentStatusCloneFailed,
	},
	DeploymentStatusBuilding: {
		DeploymentStatusStarting,
		DeploymentStatusBuildFailed,
	},
	DeploymentStatusStarting: {
		DeploymentStatusHealthChecking,
		DeploymentStatusStartFailed,
	},
	DeploymentStatusHealthChecking: {
		DeploymentStatusRunning,
		DeploymentStatusHealthCheckFailed,
	},
}

func CanTransition(from, to DeploymentStatus) bool {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}

	return false
}
