package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ProfileRepository interface {
	Create(ctx context.Context, p *dto.ProfileDTO) error
	GetByID(ctx context.Context, id int64) (*dto.ProfileDTO, error)
	List(ctx context.Context, filter dto.ListProfilesFilter) ([]dto.ProfileDTO, int, error)
	Update(ctx context.Context, id int64, req dto.UpdateProfileRequest) error
	AddSkill(ctx context.Context, profileID, skillID int64) error
	RemoveSkill(ctx context.Context, profileID, skillID int64) error
}

type OutboxRepository interface {
	Insert(ctx context.Context, eventType string, payload []byte) (string, error)
	FetchPendingBatch(ctx context.Context, limit int) ([]dto.OutboxRecord, error)
	MarkPublished(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(r Repositories) error) error
}

type Repositories interface {
	Auth() AuthRepository
	Profile() ProfileRepository
	Org() OrgRepository
	Outbox() OutboxRepository
}

type ProfileService struct {
	repo      ProfileRepository
	txManager TxManager
}

func NewProfileService(repo ProfileRepository, txManager TxManager) *ProfileService {
	return &ProfileService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *ProfileService) CreateProfile(ctx context.Context, req dto.CreateProfileRequest) (dto.ProfileDTO, error) {
	if req.Login == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.PositionID == 0 {
		return dto.ProfileDTO{}, errors.New("required fields are missing")
	}

	role := req.Role
	if role == "" {
		role = "employee"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.ProfileDTO{}, fmt.Errorf("failed to hash password: %w", err)
	}

	var createdProfile dto.ProfileDTO

	err = s.txManager.WithinTx(ctx, func(r Repositories) error {
		userID, err := r.Auth().CreateCredentials(ctx, req.Login, string(hash), role)
		if err != nil {
			return fmt.Errorf("failed to create credentials: %w", err)
		}

		profile := &dto.ProfileDTO{
			ID:         userID,
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			PositionID: req.PositionID,
			Email:      req.Email,
			AvatarURL:  req.AvatarURL,
			HireDate:   req.HireDate,
			Login:      req.Login,
			Role:       role,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if req.DepartmentID != nil {
			profile.Department = &dto.DepartmentDTO{ID: *req.DepartmentID}
		}

		if err := r.Profile().Create(ctx, profile); err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		eventPayload := dto.EmployeeEventPayload{
			EventID:   uuid.New().String(),
			UserID:    userID,
			FullName:  fmt.Sprintf("%s %s", req.FirstName, req.LastName),
			PhotoURL:  req.AvatarURL,
			Role:      role,
			Timestamp: time.Now().Unix(),
		}
		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal outbox event: %w", err)
		}

		if _, err := r.Outbox().Insert(ctx, "employee.event.created", payloadBytes); err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}

		createdProfile = *profile
		return nil
	})

	if err != nil {
		return dto.ProfileDTO{}, err
	}

	return s.GetProfile(ctx, createdProfile.ID)
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (dto.ProfileDTO, error) {
	profile, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return dto.ProfileDTO{}, err
	}
	if profile == nil {
		return dto.ProfileDTO{}, ErrNotFound
	}
	return *profile, nil
}

func (s *ProfileService) ListProfiles(ctx context.Context, filter dto.ListProfilesFilter) (dto.ListProfilesResponse, error) {
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageNumber <= 0 {
		filter.PageNumber = 1
	}

	profiles, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return dto.ListProfilesResponse{}, err
	}

	return dto.ListProfilesResponse{
		Profiles:   profiles,
		TotalCount: total,
	}, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) (dto.ProfileDTO, error) {
	err := s.txManager.WithinTx(ctx, func(r Repositories) error {
		existing, err := r.Profile().GetByID(ctx, userID)
		if err != nil || existing == nil {
			return ErrNotFound
		}

		if err := r.Profile().Update(ctx, userID, req); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}

		firstName := existing.FirstName
		if req.FirstName != nil {
			firstName = *req.FirstName
		}
		lastName := existing.LastName
		if req.LastName != nil {
			lastName = *req.LastName
		}
		avatarURL := existing.AvatarURL
		if req.AvatarURL != nil {
			avatarURL = req.AvatarURL
		}

		eventPayload := dto.EmployeeEventPayload{
			EventID:   uuid.New().String(),
			UserID:    userID,
			FullName:  fmt.Sprintf("%s %s", firstName, lastName),
			PhotoURL:  avatarURL,
			Role:      existing.Role,
			Timestamp: time.Now().Unix(),
		}
		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal outbox event: %w", err)
		}

		if _, err := r.Outbox().Insert(ctx, "employee.event.updated", payloadBytes); err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return dto.ProfileDTO{}, err
	}

	return s.GetProfile(ctx, userID)
}

func (s *ProfileService) ChangeUserStatus(ctx context.Context, userID int64, isActive bool) error {
	return s.txManager.WithinTx(ctx, func(r Repositories) error {
		if err := r.Auth().UpdateStatus(ctx, userID, isActive); err != nil {
			return err
		}

		eventType := "employee.event.updated"
		if !isActive {
			eventType = "employee.event.deleted"
		}

		eventPayload := dto.EmployeeEventPayload{
			EventID:   uuid.New().String(),
			UserID:    userID,
			Timestamp: time.Now().Unix(),
		}
		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return err
		}

		_, err = r.Outbox().Insert(ctx, eventType, payloadBytes)
		return err
	})
}

func (s *ProfileService) AddSkill(ctx context.Context, employeeID, skillID int64) error {
	return s.repo.AddSkill(ctx, employeeID, skillID)
}

func (s *ProfileService) RemoveSkill(ctx context.Context, employeeID, skillID int64) error {
	return s.repo.RemoveSkill(ctx, employeeID, skillID)
}
