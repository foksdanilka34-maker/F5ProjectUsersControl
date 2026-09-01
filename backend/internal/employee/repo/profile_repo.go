package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
)

type ProfileRepo struct {
	db DBExecutor
}

func NewProfileRepo(db DBExecutor) *ProfileRepo {
	return &ProfileRepo{db: db}
}

func (r *ProfileRepo) Create(ctx context.Context, p *dto.ProfileDTO) error {
	var hireDate time.Time
	if p.HireDate != "" {
		hireDate, _ = time.Parse("2006-01-02", p.HireDate)
	}
	if hireDate.IsZero() {
		hireDate = time.Now()
	}

	var deptID *int64
	if p.Department != nil && p.Department.ID != 0 {
		deptID = &p.Department.ID
	}

	query := `
		INSERT INTO identity.profiles (id, first_name, last_name, position_id, department_id, email, avatar_url, hire_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, p.ID, p.FirstName, p.LastName, p.PositionID, deptID, p.Email, p.AvatarURL, hireDate)
	return err
}

func (r *ProfileRepo) GetByID(ctx context.Context, id int64) (*dto.ProfileDTO, error) {
	query := `
		SELECT p.id, p.first_name, p.last_name, p.position_id, p.email, p.avatar_url, p.hire_date,
		       p.created_at, p.updated_at, c.login, c.role, c.is_active,
		       d.id, d.name
		FROM identity.profiles p
		JOIN identity.credentials c ON p.id = c.user_id
		LEFT JOIN identity.departments d ON p.department_id = d.id
		WHERE p.id = $1
	`
	var p dto.ProfileDTO
	var hireDate time.Time
	var deptID *int64
	var deptName *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.PositionID, &p.Email, &p.AvatarURL, &hireDate,
		&p.CreatedAt, &p.UpdatedAt, &p.Login, &p.Role, &p.IsActive,
		&deptID, &deptName,
	)
	if err != nil {
		return nil, err
	}
	p.HireDate = hireDate.Format("2006-01-02")
	if deptID != nil && deptName != nil {
		p.Department = &dto.DepartmentDTO{ID: *deptID, Name: *deptName}
	}

	skills, err := r.getProfileSkills(ctx, id)
	if err == nil {
		p.Skills = skills
	}

	return &p, nil
}

func (r *ProfileRepo) List(ctx context.Context, filter dto.ListProfilesFilter) ([]dto.ProfileDTO, int, error) {
	offset := (filter.PageNumber - 1) * filter.PageSize

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.DepartmentID > 0 {
		whereClause += fmt.Sprintf(" AND p.department_id = $%d", argIdx)
		args = append(args, filter.DepartmentID)
		argIdx++
	}
	if filter.PositionID > 0 {
		whereClause += fmt.Sprintf(" AND p.position_id = $%d", argIdx)
		args = append(args, filter.PositionID)
		argIdx++
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM identity.profiles p
		JOIN identity.credentials c ON p.id = c.user_id
		%s
	`, whereClause)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.first_name, p.last_name, p.position_id, p.email, p.avatar_url, p.hire_date,
		       p.created_at, p.updated_at, c.login, c.role, c.is_active,
		       d.id, d.name
		FROM identity.profiles p
		JOIN identity.credentials c ON p.id = c.user_id
		LEFT JOIN identity.departments d ON p.department_id = d.id
		%s
		ORDER BY p.id ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []dto.ProfileDTO
	for rows.Next() {
		var p dto.ProfileDTO
		var hireDate time.Time
		var deptID *int64
		var deptName *string

		if err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.PositionID, &p.Email, &p.AvatarURL, &hireDate,
			&p.CreatedAt, &p.UpdatedAt, &p.Login, &p.Role, &p.IsActive,
			&deptID, &deptName,
		); err != nil {
			return nil, 0, err
		}
		p.HireDate = hireDate.Format("2006-01-02")
		if deptID != nil && deptName != nil {
			p.Department = &dto.DepartmentDTO{ID: *deptID, Name: *deptName}
		}

		skills, _ := r.getProfileSkills(ctx, p.ID)
		p.Skills = skills

		profiles = append(profiles, p)
	}

	return profiles, total, nil
}

func (r *ProfileRepo) Update(ctx context.Context, id int64, req dto.UpdateProfileRequest) error {
	query := `
		UPDATE identity.profiles
		SET first_name = COALESCE($1, first_name),
		    last_name = COALESCE($2, last_name),
		    position_id = COALESCE($3, position_id),
		    department_id = COALESCE($4, department_id),
		    email = COALESCE($5, email),
		    avatar_url = COALESCE($6, avatar_url),
		    updated_at = NOW()
		WHERE id = $7
	`
	_, err := r.db.Exec(ctx, query, req.FirstName, req.LastName, req.PositionID, req.DepartmentID, req.Email, req.AvatarURL, id)
	return err
}

func (r *ProfileRepo) AddSkill(ctx context.Context, profileID, skillID int64) error {
	query := `
		INSERT INTO identity.profile_skills (profile_id, skill_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, profileID, skillID)
	return err
}

func (r *ProfileRepo) RemoveSkill(ctx context.Context, profileID, skillID int64) error {
	query := `DELETE FROM identity.profile_skills WHERE profile_id = $1 AND skill_id = $2`
	_, err := r.db.Exec(ctx, query, profileID, skillID)
	return err
}

func (r *ProfileRepo) getProfileSkills(ctx context.Context, profileID int64) ([]dto.SkillDTO, error) {
	query := `
		SELECT s.id, s.name
		FROM identity.skills s
		JOIN identity.profile_skills ps ON s.id = ps.skill_id
		WHERE ps.profile_id = $1
		ORDER BY s.name ASC
	`
	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []dto.SkillDTO
	for rows.Next() {
		var s dto.SkillDTO
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}
