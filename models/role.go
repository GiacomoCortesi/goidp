package models

import (
	"fmt"

	"gorm.io/gorm"
)

// RoleRepo wraps the db connection pool in a custom type
// This approach fits nicely to perform unit tests since we can reference RoleRepo in the application code
// with an interface
type RoleRepo struct {
	DB *gorm.DB
}

// UserRole represents admitted role values for a User
type UserRole int

const (
	AdminRole UserRole = iota + 1
	HelpdeskRole
	MonitorRole
)

func (r UserRole) String() string {
	return []string{"ADMIN", "HELPDESK", "MONITOR"}[r-1]
}

// Role resemble the DB roles table schema
type Role struct {
	gorm.Model
	Name  string `json:"name"`
	Users []User `gorm:"many2many:user_roles"`
}

func (r *Role) TableName() string {
	return "roles"
}

func (r *Role) ToString() string {
	return fmt.Sprintf("id: %d\nname: %s", r.ID, r.Name)
}

func (r *Role) String() string {
	return r.Name
}

func GetDefaultRoles() []Role {
	return []Role{
		{
			Name:  AdminRole.String(),
			Model: gorm.Model{ID: uint(AdminRole)},
		},
		{
			Name:  HelpdeskRole.String(),
			Model: gorm.Model{ID: uint(HelpdeskRole)},
		},
		{
			Name:  MonitorRole.String(),
			Model: gorm.Model{ID: uint(MonitorRole)},
		},
	}
}

// AddDefaultRoles add the default roles into DB
func (rR *RoleRepo) AddDefaultRoles() {
	roles := GetDefaultRoles()
	var res *gorm.DB
	for _, role := range roles {
		res = rR.DB.FirstOrCreate(&role)
	}
	if res.Error != nil {
		panic("failed to add default roles to database")
	}
}
