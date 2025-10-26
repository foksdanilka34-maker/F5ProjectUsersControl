package employee

import (
	"context"
	"errors"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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
	CreateProfile(ctx context.Context, tx pgx.Tx, regData *emp.RegisterData) (*emp.Profile, error)
	GetProfile(ctx context.Context, userID string) (*emp.Profile, error)
	ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error)
	UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error)

	CreateDepartment(ctx context.Context, name string) (*emp.Department, error)
	GetDepartment(ctx context.Context, id string) (*emp.Department, error)
	ListDepartments(ctx context.Context) ([]*emp.Department, error)
	UpdateDepartment(ctx context.Context, id, name string) (*emp.Department, error)
	DeleteDepartment(ctx context.Context, id string) error

	CreatePosition(ctx context.Context, name string) (*emp.Position, error)
	GetPosition(ctx context.Context, id string) (*emp.Position, error)
	ListPositions(ctx context.Context) ([]*emp.Position, error)
	UpdatePosition(ctx context.Context, id, name string) (*emp.Position, error)
	DeletePosition(ctx context.Context, id string) error

	CreateSkill(ctx context.Context, name string) (*emp.Skill, error)
	ListSkills(ctx context.Context) ([]*emp.Skill, error)
	AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error
	RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error

	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}

func (s *Storage) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return s.pgx.Begin(ctx)
}

func (s *Storage) CreateProfile(ctx context.Context, tx pgx.Tx, regData *emp.RegisterData) (*emp.Profile, error) {
	query := `INSERT INTO employees.profiles (first_name, last_name, position_id, email, department_id, avatar_url, hire_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at`
	profile := &emp.Profile{}

	err := tx.QueryRow(ctx, query, regData.FirstName, regData.LastName,
		regData.Position, regData.Email, regData.Departm, regData.AvatarUrl,
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
			storage.Logger.Error("no data returned from CreateProfile", zap.Error(err))
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
			storage.Logger.Warn("no profile found", zap.String("userID", userID))
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
			storage.Logger.Warn("no profiles found in ListProfile")
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
			storage.Logger.Error("error scanning profile row", zap.Error(err))
			return nil, err
		}
		data = append(data, profile)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating rows", zap.Error(err))
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
		updProf.DepartmID, updProf.AvatarUrl, updProf.HireDate, userID).Scan(
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
			storage.Logger.Warn("profile not updated", zap.String("userID", userID), zap.Error(err))
		} else {
			storage.Logger.Error("system error updating profile", zap.Error(err))
		}
		return nil, err
	}
	return profile, nil
}

func (s *Storage) CreateDepartment(ctx context.Context, name string) (*emp.Department, error) {
	query := `INSERT INTO employees.departments (name) VALUES ($1) RETURNING id, name`
	department := &emp.Department{}
	err := s.pgx.QueryRow(ctx, query, name).Scan(&department.ID, &department.Name)
	if err != nil {
		storage.Logger.Error("error creating department", zap.Error(err))
		return nil, err
	}
	return department, nil
}

func (s *Storage) GetDepartment(ctx context.Context, id string) (*emp.Department, error) {
	query := `SELECT id, name FROM employees.departments WHERE id = $1`
	department := &emp.Department{}
	err := s.pgx.QueryRow(ctx, query, id).Scan(&department.ID, &department.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("department not found", zap.String("id", id))
		} else {
			storage.Logger.Error("error getting department", zap.Error(err))
		}
		return nil, err
	}
	return department, nil
}

func (s *Storage) ListDepartments(ctx context.Context) ([]*emp.Department, error) {
	query := `SELECT id, name FROM employees.departments ORDER BY name`
	rows, err := s.pgx.Query(ctx, query)
	if err != nil {
		storage.Logger.Error("error listing departments", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	departments := make([]*emp.Department, 0)
	for rows.Next() {
		department := &emp.Department{}
		if err := rows.Scan(&department.ID, &department.Name); err != nil {
			storage.Logger.Error("error scanning department row", zap.Error(err))
			return nil, err
		}
		departments = append(departments, department)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating department rows", zap.Error(err))
		return nil, err
	}

	return departments, nil
}

func (s *Storage) UpdateDepartment(ctx context.Context, id, name string) (*emp.Department, error) {
	query := `UPDATE employees.departments SET name = $1 WHERE id = $2 RETURNING id, name`
	department := &emp.Department{}
	err := s.pgx.QueryRow(ctx, query, name, id).Scan(&department.ID, &department.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("department not found for update", zap.String("id", id))
		} else {
			storage.Logger.Error("error updating department", zap.Error(err))
		}
		return nil, err
	}
	return department, nil
}

func (s *Storage) DeleteDepartment(ctx context.Context, id string) error {
	query := `DELETE FROM employees.departments WHERE id = $1`
	cmdTag, err := s.pgx.Exec(ctx, query, id)
	if err != nil {
		storage.Logger.Error("error deleting department", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("department not found for deletion", zap.String("id", id))
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) CreatePosition(ctx context.Context, name string) (*emp.Position, error) {
	query := `INSERT INTO employees.positions (name) VALUES ($1) RETURNING id, name`
	position := &emp.Position{}
	err := s.pgx.QueryRow(ctx, query, name).Scan(&position.ID, &position.Name)
	if err != nil {
		storage.Logger.Error("error creating position", zap.Error(err))
		return nil, err
	}
	return position, nil
}

func (s *Storage) GetPosition(ctx context.Context, id string) (*emp.Position, error) {
	query := `SELECT id, name FROM employees.positions WHERE id = $1`
	position := &emp.Position{}
	err := s.pgx.QueryRow(ctx, query, id).Scan(&position.ID, &position.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("position not found", zap.String("id", id))
		} else {
			storage.Logger.Error("error getting position", zap.Error(err))
		}
		return nil, err
	}
	return position, nil
}

func (s *Storage) ListPositions(ctx context.Context) ([]*emp.Position, error) {
	query := `SELECT id, name FROM employees.positions ORDER BY name`
	rows, err := s.pgx.Query(ctx, query)
	if err != nil {
		storage.Logger.Error("error listing positions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	positions := make([]*emp.Position, 0)
	for rows.Next() {
		position := &emp.Position{}
		if err := rows.Scan(&position.ID, &position.Name); err != nil {
			storage.Logger.Error("error scanning position row", zap.Error(err))
			return nil, err
		}
		positions = append(positions, position)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating position rows", zap.Error(err))
		return nil, err
	}

	return positions, nil
}

func (s *Storage) UpdatePosition(ctx context.Context, id, name string) (*emp.Position, error) {
	query := `UPDATE employees.positions SET name = $1 WHERE id = $2 RETURNING id, name`
	position := &emp.Position{}
	err := s.pgx.QueryRow(ctx, query, name, id).Scan(&position.ID, &position.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("position not found for update", zap.String("id", id))
		} else {
			storage.Logger.Error("error updating position", zap.Error(err))
		}
		return nil, err
	}
	return position, nil
}

func (s *Storage) DeletePosition(ctx context.Context, id string) error {
	query := `DELETE FROM employees.positions WHERE id = $1`
	cmdTag, err := s.pgx.Exec(ctx, query, id)
	if err != nil {
		storage.Logger.Error("error deleting position", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("position not found for deletion", zap.String("id", id))
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) CreateSkill(ctx context.Context, name string) (*emp.Skill, error) {
	query := `INSERT INTO employees.skills (name) VALUES ($1) RETURNING id, name`
	skill := &emp.Skill{}
	err := s.pgx.QueryRow(ctx, query, name).Scan(&skill.ID, &skill.Name)
	if err != nil {
		storage.Logger.Error("error creating skill", zap.Error(err))
		return nil, err
	}
	return skill, nil
}

func (s *Storage) ListSkills(ctx context.Context) ([]*emp.Skill, error) {
	query := `SELECT id, name FROM employees.skills ORDER BY name`
	rows, err := s.pgx.Query(ctx, query)
	if err != nil {
		storage.Logger.Error("error listing skills", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	skills := make([]*emp.Skill, 0)
	for rows.Next() {
		skill := &emp.Skill{}
		if err := rows.Scan(&skill.ID, &skill.Name); err != nil {
			storage.Logger.Error("error scanning skill row", zap.Error(err))
			return nil, err
		}
		skills = append(skills, skill)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating skill rows", zap.Error(err))
		return nil, err
	}

	return skills, nil
}

func (s *Storage) AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error {
	query := `INSERT INTO employees.employee_skills (employee_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	cmdTag, err := s.pgx.Exec(ctx, query, employeeID, skillID)
	if err != nil {
		storage.Logger.Error("error adding skill to employee", zap.String("employeeID", employeeID), zap.String("skillID", skillID), zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Info("skill already assigned to employee", zap.String("employeeID", employeeID), zap.String("skillID", skillID))
	}
	return nil
}

func (s *Storage) RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error {
	query := `DELETE FROM employees.employee_skills WHERE employee_id = $1 AND skill_id = $2`
	cmdTag, err := s.pgx.Exec(ctx, query, employeeID, skillID)
	if err != nil {
		storage.Logger.Error("error removing skill from employee", zap.String("employeeID", employeeID), zap.String("skillID", skillID), zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("skill not found for employee", zap.String("employeeID", employeeID), zap.String("skillID", skillID))
		return pgx.ErrNoRows
	}
	return nil
}
