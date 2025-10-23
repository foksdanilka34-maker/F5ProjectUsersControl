package employee

import (
	"context"
	"errors"
	"fmt"
	"log"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pgx *pgxpool.Pool
}

func NewStorage(p *pgxpool.Pool) *Storage {
	return &Storage{
		pgx: p,
	}
}

type EmployeeStorage interface {
	CreateProfile(ctx context.Context, regData *emp.RegisterData) (*emp.Profile, error)
	GetProfile(ctx context.Context, userID string) (*emp.Profile, error)
	ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error)
	UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error)

	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}

func (s *Storage) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return s.pgx.Begin(ctx)
}

func (s *Storage) CreateProfile(ctx context.Context, tx pgx.Tx, regData *emp.RegisterData) (*emp.Profile, error) {
	tx, err := s.pgx.Begin(ctx)
	if err != nil {
		log.Printf("transaction err %v", err)
		return nil, fmt.Errorf("transaction err %v", err)
	}
	query := `INSERT INTO employees.profiles (first_name, last_name, position_id, email, department_id, avatar_url, hire_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at`
	profile := &emp.Profile{}

	err = tx.QueryRow(ctx, query, regData.FirstName, regData.LastName,
		regData.Position, regData.Email, regData.Departm.ID, regData.AvatarUrl,
		regData.HireDate).Scan(
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.PositionId,
		&profile.Email,
		&profile.Departm.ID,
		&profile.AvatarUrl,
		&profile.HireDate,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("error in sql expression CreateProfile, data was not returned: %v", err)
		}
		return nil, err
	}
	return profile, nil
}

func (s *Storage) GetProfile(ctx context.Context, userID string) (*emp.Profile, error) {
	query := `SELECT id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at 
			FROM employees.profiles WHERE id = $1`
	profile := &emp.Profile{}
	err := s.pgx.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.PositionId,
		&profile.Email,
		&profile.Departm.ID,
		&profile.AvatarUrl,
		&profile.HireDate,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("no data found with profile id %s", userID)
		}
		return nil, err
	}
	return profile, nil
}

func (s *Storage) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
	query := `SELECT id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at 
				FROM employees.profiles 
				WHERE ($1::UUID IS NULL OR department_id = $1) AND ($2::UUID IS NULL OR position_id = $2)
				LIMIT $3 OFFSET $4`

	rows, err := s.pgx.Query(ctx, query, departmentID, positionID, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("data not found in List method %v", err)
		}
		return nil, err
	}
	defer rows.Close()

	data := make([]*emp.Profile, 0, pageSize)
	for rows.Next() {
		profile := &emp.Profile{}
		err := rows.Scan(
			&profile.UserID,
			&profile.FirstName,
			&profile.LastName,
			&profile.PositionId,
			&profile.Email,
			&profile.Departm.ID,
			&profile.AvatarUrl,
			&profile.HireDate,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			log.Printf("error scanning profile row: %v", err)
			return nil, err
		}
		data = append(data, profile)
	}

	if err = rows.Err(); err != nil {
		log.Printf("error iterating rows: %v", err)
		return nil, err
	}

	return data, nil
}

func (s *Storage) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	query := `UPDATE employees.profiles SET
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			position_id = COALESCE($3, position_id),
			email = COALESCE($4, email),
			department_id = COALESCE($5, department_id),
			avatar_url = COALESCE($6, avatar_url),
			hire_date = COALESCE($7, hire_date),
			updated_at = NOW()
			WHERE id = $8
			RETURNING id, first_name, last_name, position_id, email, department_id, avatar_url, 
			hire_date, created_at, updated_at`

	profile := &emp.Profile{}
	err := s.pgx.QueryRow(ctx, query, updProf.FirstName, updProf.LastName, updProf.PositionId, updProf.Email,
		updProf.Departm.ID, updProf.AvatarUrl, updProf.HireDate, userID).Scan(
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.PositionId,
		&profile.Email,
		&profile.Departm.ID,
		&profile.AvatarUrl,
		&profile.HireDate,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("profile %s is not updated: %v", userID, err)
		} else {
			log.Printf("system error duting updating profile, %v", err)
		}
		return nil, err
	}
	return profile, nil
}

