package impl

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"net/url"

	kube_types "git.containerum.net/ch/kube-api/pkg/model"
	"git.containerum.net/ch/solutions/pkg/db"
	"git.containerum.net/ch/solutions/pkg/sErrors"
	"git.containerum.net/ch/solutions/pkg/server"
	"git.containerum.net/ch/solutions/pkg/utils"
	stypes "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
)

const (
	namespaceKey = "NS"
	ownerKey     = "OWNER"

	unableToCreate = "unable to create %s %s: %s"
	unableToDelete = "unable to delete %s %s: %s"
)

func (s *serverImpl) RunSolution(ctx context.Context, solutionReq stypes.UserSolution) (*stypes.RunSolutionResponce, error) {
	s.log.Infoln("Running solution ", solutionReq.Name)
	s.log.Debugln("Getting template from DB")
	solutionAvailable, err := s.svc.DB.GetTemplate(ctx, solutionReq.Template)
	if err = s.handleDBError(err); err != nil {
		return nil, err
	}

	solutionURL, err := url.Parse(solutionAvailable.URL)
	if err != nil {
		return nil, err
	}

	sName := strings.TrimSpace(solutionURL.Path[1:])

	s.log.Debugln("Downloading template config file")
	solutionF, err := s.svc.DownloadClient.DownloadFile(ctx, fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/.containerum.json", sName, solutionReq.Branch))
	if err != nil {
		return nil, err
	}

	s.log.Debugln("Generating random values for template env")
	solutionTmpl, err := template.New("solution").Funcs(template.FuncMap{
		"rand_string": utils.RandString,
	}).Parse(string(solutionF))
	if err != nil {
		return nil, err
	}

	var solutionBuf bytes.Buffer
	err = solutionTmpl.Execute(&solutionBuf, nil)
	if err != nil {
		return nil, err
	}

	var solutionConfig *server.Solution
	err = jsoniter.Unmarshal(solutionBuf.Bytes(), &solutionConfig)
	if err != nil {
		return nil, err
	}

	if len(solutionConfig.Env) == 0 {
		solutionReq.Env = make(map[string]string, 0)
	}

	s.log.Debugln("Setting envs")
	solutionConfig.Env[namespaceKey] = solutionReq.Namespace

	for k, v := range solutionReq.Env {
		solutionConfig.Env[k] = v
	}
	solutionConfig.Env[ownerKey] = httputil.MustGetUserID(ctx)

	solutionUUID := uuid.New().String()
	environments, err := jsoniter.Marshal(solutionConfig.Env)
	if err != nil {
		return nil, err
	}

	s.log.Debugln("Creating solution")
	err = s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		err = s.svc.DB.AddSolution(ctx, solutionReq, httputil.MustGetUserID(ctx), solutionAvailable.ID, solutionUUID, string(environments))
		return err
	})
	if err = s.handleDBError(err); err != nil {
		return nil, err
	}

	ret := stypes.RunSolutionResponce{
		Errors:  []string{},
		Created: 0,
	}

	s.log.Debugln("Creating solution resources")
	for _, f := range solutionConfig.Run {
		s.log.Infof("Creating %s %s", f.Type, f.Name)
		s.log.Debugln("Downloading resource")
		resF, err := s.svc.DownloadClient.DownloadFile(ctx, fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", solutionAvailable.Name, solutionReq.Branch, f.Name))
		if err != nil {
			s.log.Debugln(err)
			ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
			continue
		}

		s.log.Debugln("Setting envs to resource config")
		resTmpl, err := template.New("res").Parse(string(resF))
		if err != nil {
			s.log.Debugln(err)
			ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
			continue
		}

		var resParsed bytes.Buffer
		err = resTmpl.Execute(&resParsed, solutionConfig.Env)
		if err != nil {
			s.log.Debugln(err)
			ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
			continue
		}

		var resMetaJSON server.ResName
		err = jsoniter.Unmarshal(resParsed.Bytes(), &resMetaJSON)
		if err != nil {
			s.log.Debugln(err)
			ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
			continue
		}

		switch f.Type {
		case "deployment":
			convertedDeploy, err := s.svc.ConverterClient.ConvertDeployment(ctx, resParsed.String())
			if err != nil {
				s.log.Debugln(err)
				ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
				continue
			}

			err = s.svc.ResourceClient.CreateDeployment(ctx, solutionReq.Namespace, *convertedDeploy)
			if err != nil {
				s.log.Debugln(err)
				ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
				continue
			}

			err = s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
				err = s.svc.DB.AddDeployment(ctx, resMetaJSON.Metadata.Name, solutionUUID)
				return err
			})
			if err := s.handleDBError(err); err != nil {
				s.log.Debugln(err)
				ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
				continue
			}
			ret.Created++
		case "service":
			convertedService, err := s.svc.ConverterClient.ConvertService(ctx, resParsed.String())
			if err != nil {
				s.log.Debugln(err)
				ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
				continue
			}
			err = s.svc.ResourceClient.CreateService(ctx, solutionReq.Namespace, *convertedService)
			if err != nil {
				s.log.Debugln(err)
				ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
				continue
			}
			err = s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
				err = s.svc.DB.AddService(ctx, resMetaJSON.Metadata.Name, solutionUUID)
				return err
			})
			if err := s.handleDBError(err); err != nil {
				s.log.Debugln(err)
				ret.Errors = append(ret.Errors, fmt.Sprintf(unableToCreate, f.Type, f.Name, err))
				continue
			}
			ret.Created++
		default:
			ret.Errors = append(ret.Errors, fmt.Sprintf("Unknown resource type: %v. Skipping.", f.Type))
			continue
		}
	}

	if ret.Created == 0 {
		s.log.Infoln("No resources was created. Deleting solution...")
		err := s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
			err := s.svc.DB.DeleteSolution(ctx, solutionReq.Name, httputil.MustGetUserID(ctx))
			return err
		})
		if err != nil {
			s.log.Errorln(err)
		}
		return nil, sErrors.ErrUnableCreateSolution()
	}

	ret.NotCreated = len(ret.Errors)

	s.log.Infoln("Solution resources has been created")
	return &ret, nil
}

func (s *serverImpl) DeleteSolution(ctx context.Context, solution string) error {
	depl := make([]string, 0)
	svc := make([]string, 0)
	var ns *string

	err := s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		var err error
		depl, ns, err = s.svc.DB.GetSolutionsDeployments(ctx, solution, httputil.MustGetUserID(ctx))
		return err
	})
	if err := s.handleDBError(err); err != nil {
		return err
	}

	err = s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		var err error
		svc, _, err = s.svc.DB.GetSolutionsServices(ctx, solution, httputil.MustGetUserID(ctx))
		return err
	})
	if err := s.handleDBError(err); err != nil {
		return err
	}

	errs := []error{}
	for _, r := range depl {
		err = s.svc.ResourceClient.DeleteDeployment(ctx, *ns, r)
		if err != nil {
			errs = append(errs, fmt.Errorf(unableToDelete, "deployment", r, err))
		}
	}

	for _, r := range svc {
		err = s.svc.ResourceClient.DeleteService(ctx, *ns, r)
		if err != nil {
			errs = append(errs, fmt.Errorf(unableToDelete, "service", r, err))
		}
	}

	if len(errs) != 0 {
		return sErrors.ErrUnableDeleteSolution().AddDetailsErr(errs...)
	}

	err = s.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		var err error
		err = s.svc.DB.DeleteSolution(ctx, solution, httputil.MustGetUserID(ctx))
		return err
	})
	if err := s.handleDBError(err); err != nil {
		return err
	}

	return nil
}

func (s *serverImpl) GetSolutionsList(ctx context.Context) (*stypes.UserSolutionsList, error) {
	resp, err := s.svc.DB.GetSolutionsList(ctx, httputil.MustGetUserID(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *serverImpl) GetUserSolutionDeployments(ctx context.Context, solutionName string) (*kube_types.DeploymentsList, error) {
	depl, ns, err := s.svc.DB.GetSolutionsDeployments(ctx, solutionName, httputil.MustGetUserID(ctx))
	if err := s.handleDBError(err); err != nil {
		return nil, err
	}

	if ns == nil || len(depl) == 0 {
		return &kube_types.DeploymentsList{Deployments: make([]kube_types.DeploymentWithOwner, 0)}, nil
	}

	userdepl, err := s.svc.KubeAPIClient.GetUserDeployments(ctx, *ns, depl)
	if err != nil {
		return nil, err
	}

	return userdepl, nil
}

func (s *serverImpl) GetUserSolutionServices(ctx context.Context, solutionName string) (*kube_types.ServicesList, error) {
	svc, ns, err := s.svc.DB.GetSolutionsServices(ctx, solutionName, httputil.MustGetUserID(ctx))
	if err := s.handleDBError(err); err != nil {
		return nil, err
	}

	if ns == nil || len(svc) == 0 {
		return &kube_types.ServicesList{Services: make([]kube_types.ServiceWithOwner, 0)}, nil
	}

	usersvc, err := s.svc.KubeAPIClient.GetUserServices(ctx, *ns, svc)
	if err != nil {
		return nil, err
	}

	return usersvc, nil
}
