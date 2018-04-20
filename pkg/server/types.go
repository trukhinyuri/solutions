package server

import (
	"context"

	"io"

	stypes "git.containerum.net/ch/json-types/solutions"

	"git.containerum.net/ch/solutions/pkg/clients"
	"git.containerum.net/ch/solutions/pkg/models"
)

// SolutionsService is an interface for server "business logic"
type SolutionsService interface {
	UpdateAvailableSolutionsList(ctx context.Context) error
	GetAvailableSolutionsList(ctx context.Context) (*stypes.AvailableSolutionsList, error)
	GetAvailableSolutionEnvList(ctx context.Context, name string, branch string) (*stypes.SolutionEnv, error)
	GetAvailableSolutionResourcesList(ctx context.Context, name string, branch string) (*stypes.SolutionResources, error)
	GetUserSolutionsList(ctx context.Context) (*stypes.UserSolutionsList, error)
	DownloadSolutionConfig(ctx context.Context, solutionReq stypes.UserSolution) (solutionFile []byte, solutionName *string, err error)
	ParseSolutionConfig(ctx context.Context, solutionBody []byte, solutionReq stypes.UserSolution) (solutionConfig *Solution, solutionUUID *string, err error)
	CreateSolutionResources(ctx context.Context, solutionConfig Solution, solutionReq stypes.UserSolution, solutionName string, solutionUUID string) error
	DeleteSolution(ctx context.Context, solution string) error
	GetUserSolutionDeployments(ctx context.Context, solutionName string) (*stypes.DeploymentsList, error)
	GetUserSolutionServices(ctx context.Context, solutionName string) (*stypes.ServicesList, error)
	io.Closer
}

// Services is a collection of resources needed for server functionality.
type Services struct {
	DB              models.DB
	DownloadClient  clients.DownloadClient
	ResourceClient  clients.ResourceClient
	KubeAPIClient   clients.KubeAPIClient
	ConverterClient clients.ConverterClient
}

type Solution struct {
	Env map[string]string `json:"env"`
	Run []ConfigFile      `json:"run,omitempty"`
}

type ConfigFile struct {
	Name string `json:"config_file"`
	Type string `json:"type"`
}

type ResName struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}
