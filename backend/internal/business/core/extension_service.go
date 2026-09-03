package core

import (
	"context"
	"errors"
	"strings"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

var (
	ErrExtensionExists       = errors.New("extension with this key already exists")
	ErrExtensionNotInstalled = errors.New("extension is not installed for this project")
	ErrInvalidExtensionAuth  = errors.New("invalid extension credentials")
)

var allowedExtensionEvents = map[string]bool{
	dto.ExtEventTaskCreated:       true,
	dto.ExtEventTaskStatusChanged: true,
	dto.ExtEventCommentAdded:      true,
}

type ExtensionRepository interface {
	Create(ctx context.Context, e *dto.ExtensionDTO) (int64, error)
	GetByKey(ctx context.Context, key string) (*dto.ExtensionDTO, error)
	GetByID(ctx context.Context, id int64) (*dto.ExtensionDTO, error)
	List(ctx context.Context) ([]dto.ExtensionDTO, error)
	Delete(ctx context.Context, key string) error

	SetProjectInstall(ctx context.Context, projectID, extensionID int64, enabled bool) error
	Uninstall(ctx context.Context, projectID, extensionID int64) error
	ListForProject(ctx context.Context, projectID int64) ([]dto.ProjectExtensionDTO, error)
	IsInstalledAndEnabled(ctx context.Context, projectID, extensionID int64) (bool, error)

	SetProperty(ctx context.Context, taskID, extensionID int64, key string, value []byte) error
	GetProperty(ctx context.Context, taskID, extensionID int64, key string) ([]byte, error)
	ListProperties(ctx context.Context, taskID, extensionID int64) ([]dto.TaskEntityPropertyDTO, error)
}

type ExtensionService struct {
	repo      ExtensionRepository
	taskRepo  TaskRepository
	txManager TxManager
	sealer    TokenSealer
}

func NewExtensionService(repo ExtensionRepository, taskRepo TaskRepository, txManager TxManager, sealer TokenSealer) *ExtensionService {
	return &ExtensionService{
		repo:      repo,
		taskRepo:  taskRepo,
		txManager: txManager,
		sealer:    sealer,
	}
}

func (s *ExtensionService) Register(ctx context.Context, req dto.SaveExtensionRequest) (*dto.ExtensionDTO, error) {
	key := strings.TrimSpace(req.Key)
	if key == "" || req.Name == "" || req.BaseURL == "" || req.SharedSecret == "" {
		return nil, errors.New("key, name, base_url and shared_secret are required")
	}

	existing, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrExtensionExists
	}

	events := make([]string, 0, len(req.Events))
	for _, e := range req.Events {
		if allowedExtensionEvents[e] {
			events = append(events, e)
		}
	}

	sealed, err := s.sealer.Seal(req.SharedSecret)
	if err != nil {
		return nil, err
	}

	extension := &dto.ExtensionDTO{
		Key:             key,
		Name:            req.Name,
		Description:     req.Description,
		BaseURL:         strings.TrimRight(strings.TrimSpace(req.BaseURL), "/"),
		SharedSecretEnc: sealed,
		TaskPanelURL:    optional(req.TaskPanelURL),
		ProjectTabURL:   optional(req.ProjectTabURL),
		ProjectTabLabel: optional(req.ProjectTabLabel),
		Events:          events,
		IsActive:        true,
	}

	var id int64
	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		var err error
		id, err = r.Extension().Create(ctx, extension)
		return err
	})
	if err != nil {
		return nil, err
	}

	extension.ID = id
	return extension, nil
}

func (s *ExtensionService) List(ctx context.Context) ([]dto.ExtensionDTO, error) {
	return s.repo.List(ctx)
}

func (s *ExtensionService) Delete(ctx context.Context, key string) error {
	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.Extension().Delete(ctx, key)
	})
}

func (s *ExtensionService) SetInstalled(ctx context.Context, projectID int64, key string, enabled bool) error {
	extension, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return err
	}
	if extension == nil {
		return ErrNotFound
	}

	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.Extension().SetProjectInstall(ctx, projectID, extension.ID, enabled)
	})
}

func (s *ExtensionService) Uninstall(ctx context.Context, projectID int64, key string) error {
	extension, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return err
	}
	if extension == nil {
		return ErrNotFound
	}

	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.Extension().Uninstall(ctx, projectID, extension.ID)
	})
}

func (s *ExtensionService) ListForProject(ctx context.Context, projectID int64) ([]dto.ProjectExtensionDTO, error) {
	return s.repo.ListForProject(ctx, projectID)
}

func (s *ExtensionService) AuthenticateExtension(ctx context.Context, key, secret string) (*dto.ExtensionDTO, error) {
	extension, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if extension == nil || !extension.IsActive {
		return nil, ErrInvalidExtensionAuth
	}

	stored, err := s.sealer.Open(extension.SharedSecretEnc)
	if err != nil {
		return nil, err
	}
	if !secureEqual(stored, secret) {
		return nil, ErrInvalidExtensionAuth
	}

	return extension, nil
}

func (s *ExtensionService) SetTaskProperty(ctx context.Context, extension *dto.ExtensionDTO, taskID int64, key string, value []byte) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	installed, err := s.repo.IsInstalledAndEnabled(ctx, task.ProjectID, extension.ID)
	if err != nil {
		return err
	}
	if !installed {
		return ErrExtensionNotInstalled
	}

	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.Extension().SetProperty(ctx, taskID, extension.ID, key, value)
	})
}

func (s *ExtensionService) GetTaskProperty(ctx context.Context, extension *dto.ExtensionDTO, taskID int64, key string) ([]byte, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	installed, err := s.repo.IsInstalledAndEnabled(ctx, task.ProjectID, extension.ID)
	if err != nil {
		return nil, err
	}
	if !installed {
		return nil, ErrExtensionNotInstalled
	}

	return s.repo.GetProperty(ctx, taskID, extension.ID, key)
}

func (s *ExtensionService) ListTaskProperties(ctx context.Context, extension *dto.ExtensionDTO, taskID int64) ([]dto.TaskEntityPropertyDTO, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	installed, err := s.repo.IsInstalledAndEnabled(ctx, task.ProjectID, extension.ID)
	if err != nil {
		return nil, err
	}
	if !installed {
		return nil, ErrExtensionNotInstalled
	}

	return s.repo.ListProperties(ctx, taskID, extension.ID)
}
