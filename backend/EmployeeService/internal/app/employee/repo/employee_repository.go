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
			log.Printf("no data returned from CreateProfile: %v", err)
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
			log.Printf("no profile found: userID=%s", userID)
		}
		return nil, err
	}
	return profile, nil
}

func (s *Storage) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
	var depIDPtr *string
	if departmentID != "" {
		depIDPtr = &departmentID
	}

	var posIDPtr *string
	if positionID != "" {
		posIDPtr = &positionID
	}

	query := `SELECT id, first_name, last_name, position_id, email, department_id, avatar_url, hire_date, created_at, updated_at 
				FROM employees.profiles 
				WHERE ($1::UUID IS NULL OR department_id = $1) AND ($2::UUID IS NULL OR position_id = $2)
				LIMIT $3 OFFSET $4`

	rows, err := s.pgx.Query(ctx, query, depIDPtr, posIDPtr, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("no profiles found in ListProfile")
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
			log.Printf("profile not updated: userID=%s, %v", userID, err)
		} else {
			log.Printf("system error updating profile: %v", err)
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
		log.Printf("error creating department: %v", err)
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
			log.Printf("department not found: id=%s", id)
		} else {
			log.Printf("error getting department: %v", err)
		}
		return nil, err
	}
	return department, nil
}

func (s *Storage) ListDepartments(ctx context.Context) ([]*emp.Department, error) {
	query := `SELECT id, name FROM employees.departments ORDER BY name`
	rows, err := s.pgx.Query(ctx, query)
	if err != nil {
		log.Printf("error listing departments: %v", err)
		return nil, err
	}
	defer rows.Close()

	departments := make([]*emp.Department, 0)
	for rows.Next() {
		department := &emp.Department{}
		if err := rows.Scan(&department.ID, &department.Name); err != nil {
			log.Printf("error scanning department row: %v", err)
			return nil, err
		}
		departments = append(departments, department)
	}

	if err = rows.Err(); err != nil {
		log.Printf("error iterating department rows: %v", err)
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
			log.Printf("department not found for update: id=%s", id)
		} else {
			log.Printf("error updating department: %v", err)
		}
		return nil, err
	}
	return department, nil
}

func (s *Storage) DeleteDepartment(ctx context.Context, id string) error {
	query := `DELETE FROM employees.departments WHERE id = $1`
	cmdTag, err := s.pgx.Exec(ctx, query, id)
	if err != nil {
		log.Printf("error deleting department: %v", err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		log.Printf("department not found for deletion: id=%s", id)
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) CreatePosition(ctx context.Context, name string) (*emp.Position, error) {
	query := `INSERT INTO employees.positions (name) VALUES ($1) RETURNING id, name`
	position := &emp.Position{}
	err := s.pgx.QueryRow(ctx, query, name).Scan(&position.ID, &position.Name)
	if err != nil {
		log.Printf("error creating position: %v", err)
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
			log.Printf("position not found: id=%s", id)
		} else {
			log.Printf("error getting position: %v", err)
		}
		return nil, err
	}
	return position, nil
}

func (s *Storage) ListPositions(ctx context.Context) ([]*emp.Position, error) {
	query := `SELECT id, name FROM employees.positions ORDER BY name`
	rows, err := s.pgx.Query(ctx, query)
	if err != nil {
		log.Printf("error listing positions: %v", err)
		return nil, err
	}
	defer rows.Close()

	positions := make([]*emp.Position, 0)
	for rows.Next() {
		position := &emp.Position{}
		if err := rows.Scan(&position.ID, &position.Name); err != nil {
			log.Printf("error scanning position row: %v", err)
			return nil, err
		}
		positions = append(positions, position)
	}

	if err = rows.Err(); err != nil {
		log.Printf("error iterating position rows: %v", err)
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
			log.Printf("position not found for update: id=%s", id)
		} else {
			log.Printf("error updating position: %v", err)
		}
		return nil, err
	}
	return position, nil
}

func (s *Storage) DeletePosition(ctx context.Context, id string) error {
	query := `DELETE FROM employees.positions WHERE id = $1`
	cmdTag, err := s.pgx.Exec(ctx, query, id)
	if err != nil {
		log.Printf("error deleting position: %v", err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		log.Printf("position not found for deletion: id=%s", id)
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) CreateSkill(ctx context.Context, name string) (*emp.Skill, error) {
	query := `INSERT INTO employees.skills (name) VALUES ($1) RETURNING id, name`
	skill := &emp.Skill{}
	err := s.pgx.QueryRow(ctx, query, name).Scan(&skill.ID, &skill.Name)
	if err != nil {
		log.Printf("error creating skill: %v", err)
		return nil, err
	}
	return skill, nil
}

func (s *Storage) ListSkills(ctx context.Context) ([]*emp.Skill, error) {
	query := `SELECT id, name FROM employees.skills ORDER BY name`
	rows, err := s.pgx.Query(ctx, query)
	if err != nil {
		log.Printf("error listing skills: %v", err)
		return nil, err
	}
	defer rows.Close()

	skills := make([]*emp.Skill, 0)
	for rows.Next() {
		skill := &emp.Skill{}
		if err := rows.Scan(&skill.ID, &skill.Name); err != nil {
			log.Printf("error scanning skill row: %v", err)
			return nil, err
		}
		skills = append(skills, skill)
	}

	if err = rows.Err(); err != nil {
		log.Printf("error iterating skill rows: %v", err)
		return nil, err
	}

	return skills, nil
}

func (s *Storage) AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error {
	query := `INSERT INTO employees.employee_skills (employee_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	cmdTag, err := s.pgx.Exec(ctx, query, employeeID, skillID)
	if err != nil {
		log.Printf("error adding skill to employee: employeeID=%s, skillID=%s, %v", employeeID, skillID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		log.Printf("skill already assigned to employee: employeeID=%s, skillID=%s", employeeID, skillID)
	}
	return nil
}

func (s *Storage) RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error {
	query := `DELETE FROM employees.employee_skills WHERE employee_id = $1 AND skill_id = $2`
	cmdTag, err := s.pgx.Exec(ctx, query, employeeID, skillID)
	if err != nil {
		log.Printf("error removing skill from employee: employeeID=%s, skillID=%s, %v", employeeID, skillID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		log.Printf("skill not found for employee: employeeID=%s, skillID=%s", employeeID, skillID)
		return pgx.ErrNoRows
	}
	return nil
}
