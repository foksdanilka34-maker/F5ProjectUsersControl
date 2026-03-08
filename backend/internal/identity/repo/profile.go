package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type ProfileRepo struct {
	pool *pgxpool.Pool
}

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) Create(ctx context.Context, tx pgx.Tx, p *Profile) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO identity.profiles (id, first_name, last_name, position_id, department_id, email, avatar_url, hire_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, p.ID, p.FirstName, p.LastName, p.PositionID, p.DepartmentID, p.Email, p.AvatarURL, p.HireDate, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *ProfileRepo) GetByID(ctx context.Context, id int64) (*Profile, error) {
	p := &Profile{}
	var deptID *int64
	var deptName *string

	err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.first_name, p.last_name, p.position_id, p.department_id, p.email, 
		       COALESCE(p.avatar_url, ''), p.hire_date, p.created_at, p.updated_at,
		       c.login, c.role, c.is_active,
		       d.id, d.name
		FROM identity.profiles p
		JOIN identity.credentials c ON c.user_id = p.id
		LEFT JOIN identity.departments d ON d.id = p.department_id
		WHERE p.id = $1
	`, id).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.PositionID, &p.DepartmentID, &p.Email,
		&p.AvatarURL, &p.HireDate, &p.CreatedAt, &p.UpdatedAt,
		&p.Login, &p.Role, &p.IsActive,
		&deptID, &deptName,
	)
	if err != nil {
		return nil, err
	}

	if deptID != nil && deptName != nil {
		p.Department = &Department{ID: *deptID, Name: *deptName}
	}

	p.Skills, _ = r.GetProfileSkills(ctx, id)

	return p, nil
}

func (r *ProfileRepo) List(ctx context.Context, pageSize, offset int, departmentID, positionID int64) ([]*Profile, int, error) {

	countBuilder := psql.Select("COUNT(*)").
		From("identity.profiles p").
		Join("identity.credentials c ON c.user_id = p.id")

	if departmentID > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"p.department_id": departmentID})
	}
	if positionID > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"p.position_id": positionID})
	}

	countQuery, countArgs, _ := countBuilder.ToSql()
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryBuilder := psql.Select(
		"p.id", "p.first_name", "p.last_name", "p.position_id", "p.department_id", "p.email",
		"COALESCE(p.avatar_url, '')", "p.hire_date", "p.created_at", "p.updated_at",
		"c.login", "c.role", "c.is_active",
		"d.id", "d.name",
	).
		From("identity.profiles p").
		Join("identity.credentials c ON c.user_id = p.id").
		LeftJoin("identity.departments d ON d.id = p.department_id")

	if departmentID > 0 {
		queryBuilder = queryBuilder.Where(sq.Eq{"p.department_id": departmentID})
	}
	if positionID > 0 {
		queryBuilder = queryBuilder.Where(sq.Eq{"p.position_id": positionID})
	}

	queryBuilder = queryBuilder.OrderBy("p.created_at DESC").Limit(uint64(pageSize)).Offset(uint64(offset))

	query, args, _ := queryBuilder.ToSql()
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	profiles := []*Profile{}
	for rows.Next() {
		p := &Profile{}
		var deptID *int64
		var deptName *string

		if err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.PositionID, &p.DepartmentID, &p.Email,
			&p.AvatarURL, &p.HireDate, &p.CreatedAt, &p.UpdatedAt,
			&p.Login, &p.Role, &p.IsActive,
			&deptID, &deptName,
		); err != nil {
			return nil, 0, err
		}

		if deptID != nil && deptName != nil {
			p.Department = &Department{ID: *deptID, Name: *deptName}
		}

		profiles = append(profiles, p)
	}

	return profiles, total, nil
}

func (r *ProfileRepo) Update(ctx context.Context, id int64, firstName, lastName *string, positionID, departmentID *int64, email, avatarURL *string) error {
	builder := psql.Update("identity.profiles").Set("updated_at", time.Now())

	if firstName != nil {
		builder = builder.Set("first_name", *firstName)
	}
	if lastName != nil {
		builder = builder.Set("last_name", *lastName)
	}
	if positionID != nil {
		builder = builder.Set("position_id", *positionID)
	}
	if departmentID != nil {
		builder = builder.Set("department_id", *departmentID)
	}
	if email != nil {
		builder = builder.Set("email", *email)
	}
	if avatarURL != nil {
		builder = builder.Set("avatar_url", *avatarURL)
	}

	builder = builder.Where(sq.Eq{"id": id})

	query, args, _ := builder.ToSql()
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *ProfileRepo) CreateDepartment(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `INSERT INTO identity.departments (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	return id, err
}

func (r *ProfileRepo) GetDepartment(ctx context.Context, id int64) (*Department, error) {
	d := &Department{}
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM identity.departments WHERE id = $1`, id).Scan(&d.ID, &d.Name)
	return d, err
}

func (r *ProfileRepo) ListDepartments(ctx context.Context) ([]*Department, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM identity.departments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	departments := []*Department{}
	for rows.Next() {
		d := &Department{}
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, nil
}

func (r *ProfileRepo) UpdateDepartment(ctx context.Context, id int64, name string) error {
	_, err := r.pool.Exec(ctx, `UPDATE identity.departments SET name = $1 WHERE id = $2`, name, id)
	return err
}

func (r *ProfileRepo) DeleteDepartment(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identity.departments WHERE id = $1`, id)
	return err
}

func (r *ProfileRepo) CreatePosition(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `INSERT INTO identity.positions (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	return id, err
}

func (r *ProfileRepo) GetPosition(ctx context.Context, id int64) (*Position, error) {
	p := &Position{}
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM identity.positions WHERE id = $1`, id).Scan(&p.ID, &p.Name)
	return p, err
}

func (r *ProfileRepo) ListPositions(ctx context.Context) ([]*Position, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM identity.positions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []*Position{}
	for rows.Next() {
		p := &Position{}
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, nil
}

func (r *ProfileRepo) UpdatePosition(ctx context.Context, id int64, name string) error {
	_, err := r.pool.Exec(ctx, `UPDATE identity.positions SET name = $1 WHERE id = $2`, name, id)
	return err
}

func (r *ProfileRepo) DeletePosition(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identity.positions WHERE id = $1`, id)
	return err
}

func (r *ProfileRepo) CreateSkill(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `INSERT INTO identity.skills (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	return id, err
}

func (r *ProfileRepo) ListSkills(ctx context.Context) ([]Skill, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM identity.skills ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []Skill{}
	for rows.Next() {
		s := Skill{}
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *ProfileRepo) DeleteSkill(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identity.skills WHERE id = $1`, id)
	return err
}

func (r *ProfileRepo) GetProfileSkills(ctx context.Context, profileID int64) ([]Skill, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name 
		FROM identity.skills s 
		JOIN identity.profile_skills ps ON ps.skill_id = s.id 
		WHERE ps.profile_id = $1
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []Skill{}
	for rows.Next() {
		s := Skill{}
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *ProfileRepo) AddSkillToProfile(ctx context.Context, profileID, skillID int64) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO identity.profile_skills (profile_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, profileID, skillID)
	return err
}

func (r *ProfileRepo) RemoveSkillFromProfile(ctx context.Context, profileID, skillID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identity.profile_skills WHERE profile_id = $1 AND skill_id = $2`, profileID, skillID)
	return err
}

func (r *ProfileRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}


