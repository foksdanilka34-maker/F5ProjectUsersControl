package employee

import (
	"time"
)

type Profile struct {
	UserID     string
	FirstName  string
	LastName   string
	PositionId string
	Email      string
	AvatarUrl  string
	HireDate   time.Time
	Departm    Department
	Login      string
	Password   string
	Role       string
	Skill      Skill

	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

type RegisterData struct {
	FirstName string
	LastName  string
	Position  string
	Email     string
	AvatarUrl string
	HireDate  time.Time
	Departm   *string
	Login     string
	Password  string
	Role      string
}

type Department struct {
	ID   string
	Name string
}

type Skill struct {
	ID   string
	Name string
}

type UpdateProfile struct {
	FirstName  *string
	LastName   *string
	PositionId *string
	Email      *string
	AvatarUrl  *string
	HireDate   *string
	DepartmID  *string

	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}