package repo

import (
	"context"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
)

type OrgRepo struct {
	db DBExecutor
}

func NewOrgRepo(db DBExecutor) *OrgRepo {
	return &OrgRepo{db: db}
}

// Departments
func (r *OrgRepo) CreateDepartment(ctx context.Context, name string) (*dto.DepartmentDTO, error) {
	query := `INSERT INTO identity.departments (name, created_at) VALUES ($1, NOW()) RETURNING id, name`
	var d dto.DepartmentDTO
	err := r.db.QueryRow(ctx, query, name).Scan(&d.ID, &d.Name)
	return &d, err
}

func (r *OrgRepo) GetDepartment(ctx context.Context, id int64) (*dto.DepartmentDTO, error) {
	query := `SELECT id, name FROM identity.departments WHERE id = $1`
	var d dto.DepartmentDTO
	err := r.db.QueryRow(ctx, query, id).Scan(&d.ID, &d.Name)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *OrgRepo) ListDepartments(ctx context.Context) ([]dto.DepartmentDTO, error) {
	query := `SELECT id, name FROM identity.departments ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.DepartmentDTO
	for rows.Next() {
		var d dto.DepartmentDTO
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *OrgRepo) UpdateDepartment(ctx context.Context, id int64, name string) (*dto.DepartmentDTO, error) {
	query := `UPDATE identity.departments SET name = $1 WHERE id = $2 RETURNING id, name`
	var d dto.DepartmentDTO
	err := r.db.QueryRow(ctx, query, name, id).Scan(&d.ID, &d.Name)
	return &d, err
}

func (r *OrgRepo) DeleteDepartment(ctx context.Context, id int64) error {
	query := `DELETE FROM identity.departments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Positions
func (r *OrgRepo) CreatePosition(ctx context.Context, name string) (*dto.PositionDTO, error) {
	query := `INSERT INTO identity.positions (name, created_at) VALUES ($1, NOW()) RETURNING id, name`
	var p dto.PositionDTO
	err := r.db.QueryRow(ctx, query, name).Scan(&p.ID, &p.Name)
	return &p, err
}

func (r *OrgRepo) GetPosition(ctx context.Context, id int64) (*dto.PositionDTO, error) {
	query := `SELECT id, name FROM identity.positions WHERE id = $1`
	var p dto.PositionDTO
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *OrgRepo) ListPositions(ctx context.Context) ([]dto.PositionDTO, error) {
	query := `SELECT id, name FROM identity.positions ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.PositionDTO
	for rows.Next() {
		var p dto.PositionDTO
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *OrgRepo) UpdatePosition(ctx context.Context, id int64, name string) (*dto.PositionDTO, error) {
	query := `UPDATE identity.positions SET name = $1 WHERE id = $2 RETURNING id, name`
	var p dto.PositionDTO
	err := r.db.QueryRow(ctx, query, name, id).Scan(&p.ID, &p.Name)
	return &p, err
}

func (r *OrgRepo) DeletePosition(ctx context.Context, id int64) error {
	query := `DELETE FROM identity.positions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Skills
func (r *OrgRepo) CreateSkill(ctx context.Context, name string) (*dto.SkillDTO, error) {
	query := `INSERT INTO identity.skills (name, created_at) VALUES ($1, NOW()) RETURNING id, name`
	var s dto.SkillDTO
	err := r.db.QueryRow(ctx, query, name).Scan(&s.ID, &s.Name)
	return &s, err
}

func (r *OrgRepo) ListSkills(ctx context.Context) ([]dto.SkillDTO, error) {
	query := `SELECT id, name FROM identity.skills ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.SkillDTO
	for rows.Next() {
		var s dto.SkillDTO
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *OrgRepo) DeleteSkill(ctx context.Context, id int64) error {
	query := `DELETE FROM identity.skills WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
