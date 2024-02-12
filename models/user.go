package models

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepo wraps the db connection pool in a custom type
// This approach fits nicely to perform unit tests since we can reference UserRepo in the application code
// with an interface
type UserRepo struct {
	DB *gorm.DB
}

// RoleList defines the list of roles assigned to a DB user
type RoleList []Role

// NewRoleList convert an array of strings into a RoleList array
func NewRoleList(roleListString []string) (RoleList, error) {
	var rL RoleList
	for _, rS := range roleListString {
		var r Role
		switch strings.ToLower(rS) {
		case strings.ToLower(AdminRole.String()):
			r.ID = uint(AdminRole)
		case strings.ToLower(HelpdeskRole.String()):
			r.ID = uint(HelpdeskRole)
		case strings.ToLower(MonitorRole.String()):
			r.ID = uint(MonitorRole)
		default:
			return nil, fmt.Errorf("invalid role %s", rS)
		}
		r.Name = rS
		rL = append(rL, r)
	}
	return rL, nil
}

func (rL *RoleList) String() []string {
	var roleListString []string
	for _, r := range *rL {
		roleListString = append(roleListString, r.String())
	}
	return roleListString
}

// User resemble the DB users table schema
// it has a many to many relationships with user roles
// each user can be assigned to multiple roles
// each role can be assigned to multiple users
type User struct {
	gorm.Model
	Username string
	Password string
	Version  int
	Roles    RoleList `gorm:"many2many:user_roles"`
}

// TableName returns the User table name
func (u *User) TableName() string {
	return "users"
}

// ToString provides a string representation of the User information
func (u *User) ToString() string {
	return fmt.Sprintf("id: %d\nusername: %s\nversion: %d\nroles: %s", u.ID, u.Username, u.Version, u.Roles.String())
}

// Create creates a new user into the DB
// It returns an error if:
// - username is already present in database
// - ID is already present in database
// - password does not satisfy security requirements
func (uR *UserRepo) Create(u *User) error {
	if u.ID != 0 {
		res := uR.DB.First(&u, u.ID)
		if res.RowsAffected != 0 {
			return &UserError{fmt.Sprintf("user ID %d already present in database", u.ID)}
		}
	}
	if err := uR.ValidateUsername(u.Username); err != nil {
		return &UserError{err.Error()}
	}
	if err := validatePassword(u.Password); err != nil {
		return &UserError{fmt.Sprintf("password does not meet security requirements: %s", err.Error())}
	}
	var res *gorm.DB

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 8)
	u.Password = string(hashedPassword)

	res = uR.DB.Create(&u)
	if res.Error != nil {
		return &DBError{res.Error.Error()}
	}
	return nil
}

// ValidateUsername validates the user username
func (uR *UserRepo) ValidateUsername(u string) error {
	var users []User
	if u == "" {
		return errors.New("username cannot be empty")
	}
	var res *gorm.DB
	res = uR.DB.Where("username = ?", u).Find(&users)
	if res.RowsAffected > 0 {
		return errors.New(fmt.Sprintf("user %s already present in database", u))
	}
	return nil
}

// UpdateUserByNameOrID update User information given user ID or user name
// it is considered to be an ID, if it can be converted into integer, otherwise
// it is considered a username
func (uR *UserRepo) UpdateUserByNameOrID(nameOrID, username, password string, roles []string) (*User, error) {
	targetUser := &User{}
	if id, err := strconv.Atoi(nameOrID); err == nil {
		targetUser.Model = gorm.Model{ID: uint(id)}
	} else {
		targetUser.Username = nameOrID
	}

	u, err := uR.GetUserByNameOrID(nameOrID)
	if err != nil {
		return nil, err
	}

	rL, err := NewRoleList(roles)
	if err != nil {
		return nil, &UserError{error: err.Error()}
	}
	if len(rL) > 0 {
		u.Roles = rL
	}
	if username != "" {
		u.Username = username
	}
	if password != "" {
		if err := validatePassword(password); err != nil {
			return nil, &UserError{fmt.Sprintf("password does not meet security requirements: %s", err.Error())}
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 8)
		u.Password = string(hashedPassword)
	}
	u.Version = u.Version + 1

	id, err := strconv.Atoi(nameOrID)
	if err == nil {
		u.ID = uint(id)
	}
	if err = uR.UpdateUser(u); err != nil {
		return nil, err
	}
	return u, err
}

// UpdateUser updates user information into the DB
func (uR *UserRepo) UpdateUser(u *User) error {
	if u.ID == 0 {
		return &UserError{"user ID is required"}
	}
	res := uR.DB.Updates(u)
	if res.Error != nil {
		return &DBError{res.Error.Error()}
	}
	var roles RoleList
	for _, r := range u.Roles {
		roles = append(roles, r)
	}
	if err := uR.DB.Model(&u).Association("Roles").Replace(&roles); err != nil {
		return &DBError{res.Error.Error()}
	}
	return nil
}

// GetUserByNameOrID retrieves user information by ID or username
func (uR *UserRepo) GetUserByNameOrID(nameOrID string) (*User, error) {
	user := &User{}
	var res *gorm.DB
	if id, err := strconv.Atoi(nameOrID); err == nil {
		res = uR.DB.Preload("Roles").First(&user, id)
	} else {
		res = uR.DB.Preload("Roles").Where("username = ?", nameOrID).First(&user)
	}
	switch res.RowsAffected {
	case -1:
		return nil, &DBError{error: res.Error.Error()}
	case 0:
		return nil, &NotFoundError{error: fmt.Sprintf("user %s not present in database", nameOrID)}
	}

	return user, nil
}

// GetUsers returns the list of User present in DB
func (uR *UserRepo) GetUsers() ([]*User, error) {
	var users []*User
	// TODO: Query to all the users won't return the Roles information
	// there for sure is a way to do that, without running a query to each user individually
	//var user User
	//err := GetDB().Table(user.TableName()).Preload("Roles").Find(&users).Error
	//if err != nil {
	//	return nil, &DBError{err.Error()}
	//}
	usersID := uR.GetUsersID()
	for _, id := range usersID {
		u, _ := uR.GetUserByNameOrID(strconv.Itoa(int(id)))
		users = append(users, u)
	}
	return users, nil
}

// GetUsersID returns the list of User IDs present in DB
func (uR *UserRepo) GetUsersID() []uint {
	var users []*User
	var user User
	uR.DB.Table(user.TableName()).Select("id").Scan(&users)
	var usersID []uint
	for _, u := range users {
		usersID = append(usersID, u.ID)
	}
	return usersID
}

// DeleteUser remove specified user from DB
func (uR *UserRepo) DeleteUser(user *User) error {
	res := uR.DB.Unscoped().Select("Roles").Delete(user)
	if res.Error != nil {
		return &DBError{res.Error.Error()}
	}
	if res.RowsAffected == 0 {
		return &NotFoundError{fmt.Sprintf("user %s not present in database", user.ToString())}
	}
	return nil
}

// DeleteUserByID remove user from DB given user ID
func (uR *UserRepo) DeleteUserByID(id int) error {
	if id == 1 {
		return &UserError{"user with id 1 cannot be deleted"}
	}
	return uR.DeleteUser(&User{Model: gorm.Model{ID: uint(id)}})
}

// DeleteUserByName remove user from DB given user name
func (uR *UserRepo) DeleteUserByName(name string) error {
	return uR.DeleteUser(&User{Username: name})
}

// DeleteUserByNameOrID remove user from DB given username or ID
// it is considered to be an ID, if it can be converted into integer, otherwise
// it is considered a username
func (uR *UserRepo) DeleteUserByNameOrID(nameOrID string) error {
	if id, err := strconv.Atoi(nameOrID); err == nil {
		return uR.DeleteUserByID(id)
	}
	return uR.DeleteUserByName(nameOrID)
}

// GetDefaultUser returns default admin user
func GetDefaultUser() *User {
	return &User{
		Username: "admin",
		Password: "admin",
		Model:    gorm.Model{ID: 1},
		Roles: []Role{
			{
				Model: gorm.Model{ID: uint(AdminRole)},
			},
		},
		Version: 1,
	}
}

// AddDefaultUser add default admin user to DB
func (uR *UserRepo) AddDefaultUser() {
	defaultUser := GetDefaultUser()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(defaultUser.Password), 8)
	defaultUser.Password = string(hashedPassword)
	var res *gorm.DB
	res = uR.DB.FirstOrCreate(&defaultUser)
	if res.Error != nil {
		panic("failed to add default user to database")
	}
}

// DeleteAllUsers remove all Users from User table
func (uR *UserRepo) DeleteAllUsers() error {
	usersID := uR.GetUsersID()
	for _, id := range usersID {
		_ = uR.DeleteUserByID(int(id))
	}
	return nil
}

func (uR *UserRepo) GetAndValidateUser(username string, password string) (*User, bool) {
	storedUser, err := uR.GetUserByNameOrID(username)
	if err != nil {
		return nil, false
	}
	if err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(password)); err != nil {
		return nil, false
	}
	return storedUser, true
}

// validatePassword make sure that provided password is strong enough
func validatePassword(p string) error {
	matchLower := regexp.MustCompile(`[a-z]`)
	matchUpper := regexp.MustCompile(`[A-Z]`)
	matchNumber := regexp.MustCompile(`[0-9]`)
	matchSpecial := regexp.MustCompile(`[\!\@\#\$\%\^\&\*\(\\\)\-_\=\+\,\.\?\/\:\;\{\}\[\]~]`)

	if len(p) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if !matchLower.MatchString(p) {
		return errors.New("password must contain a lower character")
	}
	if !matchUpper.MatchString(p) {
		return errors.New("password must contain an upper character")
	}
	if !matchNumber.MatchString(p) {
		return errors.New("password must contain a digit")
	}
	if !matchSpecial.MatchString(p) {
		return errors.New("password must contain a special character")
	}
	return nil
}
