package controllers

import (
	"fmt"
	"github.com/goidp/models"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// UsersHandler is the function which verifies what action to take based on the request (if GET or POST)
// This handler is called when no ID is passed into the http incoming request
func (a *App) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.GetUsersHandler(w, r)
	case http.MethodPost:
		a.CreateUserHandler(w, r)
	}
}

type UserResponseList []*UserResponse

type UserResponse struct {
	ID       uint     `jsonapi:"primary,user,omitempty" json:"id,omitempty"`
	Roles    []string `jsonapi:"attr,roles" json:"roles,omitempty"`
	Username string   `jsonapi:"attr,username" json:"username,omitempty"`
	Version  int      `jsonapi:"attr,version" json:"version,omitempty"`
}

type SessionExpire time.Time

func (uL SessionExpire) JSONAPIMeta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"session_expires": uL,
	}
}

type UserRequest struct {
	ID       string   `jsonapi:"primary,user,omitempty"`
	Roles    []string `jsonapi:"attr,roles"`
	Username string   `jsonapi:"attr,username"`
	Password string   `jsonapi:"attr,password"`
	Version  int      `jsonapi:"attr,version"`
}

func (a *App) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// read http request body
	var requestBody UserRequest
	if err := jsonapi.UnmarshalPayload(r.Body, &requestBody); err != nil {
		jsonapiError(w, http.StatusBadRequest, err.Error())
		return
	}
	// convert user info from request to db format
	rL, err := models.NewRoleList(requestBody.Roles)
	if err != nil {
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user := models.User{
		Username: requestBody.Username,
		Password: requestBody.Password,
		Roles:    rL,
		Version:  1,
	}
	// not really used, but we could create a user with a specific ID
	if requestBody.ID != "" {
		// if the ID doesn't parse to int, then we just ignore it
		id, _ := strconv.Atoi(requestBody.ID)
		user.Model = gorm.Model{ID: uint(id)}
	}
	// create db user
	switch err := a.Users.Create(&user).(type) {
	case *models.UserError:
		jsonapiError(w, http.StatusBadRequest, err.Error())
		return
	case *models.DBError:
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// generate user creation event
	err = a.Events.CreateUserEvent(r.Method, requestBody.Username, "")
	if err != nil {
		log.WithError(err).Warnf("failed to store user event")
	}

	// write http response
	uR := UserResponse{
		ID:       user.ID,
		Roles:    user.Roles.String(),
		Username: user.Username,
		Version:  user.Version,
	}
	jsonapiSuccess(w, &uR, http.StatusOK)
}

func (a *App) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := a.Users.GetUsers()
	var userResponseList UserResponseList
	for _, u := range users {
		userResponseList = append(userResponseList, &UserResponse{
			ID:       u.ID,
			Roles:    u.Roles.String(),
			Username: u.Username,
			Version:  u.Version,
		})
	}
	if err != nil {
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var sE SessionExpire
	jsonapiSuccessWithMeta(w, sE, userResponseList, http.StatusOK)
}

// UserHandler is the function which verifies what action to take based on the request (if GET, PATCH or DELETE)
// This handler is called when user ID is passed into the http incoming request
func (a *App) UserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		a.PatchUserHandler(w, r, id)
	case http.MethodDelete:
		a.DeleteUserHandler(w, r, id)
	case http.MethodGet:
		a.GetUserHandler(w, r, id)
	}
}

func (a *App) GetUserHandler(w http.ResponseWriter, r *http.Request, id string) {
	user, err := a.Users.GetUserByNameOrID(id)
	if err != nil {
		jsonapiError(w, http.StatusNotFound, fmt.Sprintf("user %s is not found", id))
		return
	}
	uR := &UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Version:  user.Version,
	}
	if user.Roles != nil {
		uR.Roles = user.Roles.String()
	}
	var sE SessionExpire
	jsonapiSuccessWithMeta(w, sE, uR, http.StatusOK)
}

func (a *App) PatchUserHandler(w http.ResponseWriter, r *http.Request, id string) {
	// the frontend is using username for PATCH

	var user *models.User

	u, err := a.Users.GetUserByNameOrID(id)
	if err == nil {
		// read http request body
		var requestBody UserRequest
		if err = jsonapi.UnmarshalPayload(r.Body, &requestBody); err != nil {
			jsonapiError(w, http.StatusBadRequest, err.Error())
			return
		}
		user, err = a.Users.UpdateUserByNameOrID(id, requestBody.Username, requestBody.Password, requestBody.Roles)
	}

	switch err.(type) {
	case *models.NotFoundError:
		jsonapiError(w, http.StatusNotFound, err.Error())
		return
	case *models.UserError:
		jsonapiError(w, http.StatusBadRequest, err.Error())
		return
	case *models.DBError:
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// generate user update event
	err = a.Events.CreateUserEvent(r.Method, u.Username, "")
	if err != nil {
		log.WithError(err).Warnf("failed to store user event")
	}

	userResponse := &UserResponse{
		ID:       user.ID,
		Roles:    user.Roles.String(),
		Username: user.Username,
		Version:  user.Version,
	}
	jsonapiSuccess(w, userResponse, http.StatusOK)
}

func (a *App) DeleteUserHandler(w http.ResponseWriter, r *http.Request, id string) {
	// the frontend is using ID for DELETE
	u, err := a.Users.GetUserByNameOrID(id)
	if err == nil {
		err = a.Users.DeleteUser(u)
	}
	switch err.(type) {
	case *models.NotFoundError:
		jsonapiError(w, http.StatusNotFound, fmt.Sprintf("user %s not found in database", id))
		return
	case *models.UserError:
		jsonapiError(w, http.StatusBadRequest, err.Error())
		return
	case *models.DBError:
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// generate user deletion event
	err = a.Events.CreateUserEvent(r.Method, u.Username, "")
	if err != nil {
		log.WithError(err).Warnf("failed to store user event")
	}
	jsonapiNoContentSuccess(w)
}
