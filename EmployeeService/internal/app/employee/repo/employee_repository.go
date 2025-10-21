package employee

import (
	"context"
	"errors"
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
	UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*emp.Profile, error)
	DeactivateProfile(ctx context.Context, userID string) error
}

func (s *Storage) CreateProfile(ctx context.Context, regData *emp.RegisterData) (*emp.Profile, error) {
	query := `INSERT INTO employees.profiles (first_name, last_name, position_id, email, department_id, avatar_url, hire_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at`
	profile := &emp.Profile{}

	err := s.pgx.QueryRow(ctx, query, regData.FirstName, regData.LastName,
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

func (s *Storage) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*emp.Profile, error) {
	query := `UPDATE employees.profiles SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1
	
	if firstName, ok := updates["first_name"]; ok {
		query += `, first_name = $` + string(rune(argIdx))
		args = append(args, firstName)
		argIdx++
	}
	if lastName, ok := updates["last_name"]; ok {
		query += `, last_name = $` + string(rune(argIdx))
		args = append(args, lastName)
		argIdx++
	}
	if positionID, ok := updates["position_id"]; ok {
		query += `, position_id = $` + string(rune(argIdx))
		args = append(args, positionID)
		argIdx++
	}
	if email, ok := updates["email"]; ok {
		query += `, email = $` + string(rune(argIdx))
		args = append(args, email)
		argIdx++
	}
	if departmentID, ok := updates["department_id"]; ok {
		query += `, department_id = $` + string(rune(argIdx))
		args = append(args, departmentID)
		argIdx++
	}
	if avatarURL, ok := updates["avatar_url"]; ok {
		query += `, avatar_url = $` + string(rune(argIdx))
		args = append(args, avatarURL)
		argIdx++
	}
	
	query += ` WHERE id = $` + string(rune(argIdx)) + ` RETURNING id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at`
	args = append(args, userID)
	
	profile := &emp.Profile{}
	err := s.pgx.QueryRow(ctx, query, args...).Scan(
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
			log.Printf("profile not found for update: %s", userID)
		}
		return nil, err
	}
	return profile, nil
}

func (s *Storage) DeactivateProfile(ctx context.Context, userID string) error {
	query := `UPDATE employees.profiles SET updated_at = NOW() WHERE id = $1`
	
	result, err := s.pgx.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("error deactivating profile: %v", err)
		return err
	}
	
	if result.RowsAffected() == 0 {
		log.Printf("profile not found for deactivation: %s", userID)
		return errors.New("profile not found")
	}
	
	return nil
}

