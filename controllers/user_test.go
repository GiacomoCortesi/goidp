package controllers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/goidp/models"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestCreateUsersHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	tt := []struct {
		name     string
		username string
		password string
		status   int
		err      string
	}{
		{
			name:     "empty credentials",
			username: "",
			password: "",
			status:   400,
			err:      "username cannot be empty",
		},
		{
			name:     "wrong credentials",
			username: "asdf",
			password: "qwerty",
			status:   400,
			err:      "password does not meet security requirements: password must be at least 8 characters long",
		},
		{
			name:     "correct credentials",
			username: "admin",
			password: "AdminUser1*",
			status:   200,
		},
	}

	rL, _ := models.NewRoleList([]string{models.AdminRole.String()})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	adminUser := &models.User{
		Username: "admin",
		Version:  1,
		Roles:    rL,
	}
	adminUser.Password = string(hashedPassword)

	a := NewApp(s.DB, &Config{})
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			p := &UserRequest{
				Username: tc.username,
				Password: tc.password,
			}

			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, p); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("POST", "/user", requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()

			if tc.status == http.StatusOK {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL`)).
					WithArgs(tc.username).WillReturnRows(sqlmock.NewRows(nil))

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				insertArgsUsers := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}
				insertArgsEvents := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password","version") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).
					WithArgs(insertArgsUsers...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "events" ("created_at","updated_at","deleted_at","username","activated","description","modified","authn_domain","severity") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WithArgs(insertArgsEvents...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				s.mock.ExpectCommit()
			} else if tc.name == "wrong credentials" {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL`)).
					WithArgs(tc.username).WillReturnRows(sqlmock.NewRows(nil))
			}

			a.UsersHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("expected status %d; got %d", tc.status, res.StatusCode)
			}

			if tc.err != "" {
				var uR jsonapi.ErrorsPayload
				if err = json.NewDecoder(res.Body).Decode(&uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
				if uR.Errors[0].Detail != tc.err {
					t.Fatalf("expected error detail: %s, got: %s", uR.Errors[0].Detail, tc.err)
				}
			} else {
				var uR UserResponse
				if err = jsonapi.UnmarshalPayload(res.Body, &uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
			}
		})

	}
}

func TestGetUsersHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	tt := []struct {
		name     string
		username string
		password string
		status   int
		err      string
	}{
		{
			name:     "correct credentials",
			username: "admin",
			password: "AdminUser1*",
			status:   200,
		},
	}

	a := NewApp(s.DB, &Config{})

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			p := &UserRequest{
				Username: tc.username,
				Password: tc.password,
			}

			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, p); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("GET", "/user", requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()

			s.mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT id FROM "users"`)).
				WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			a.UsersHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("expected status %d; got %d", tc.status, res.StatusCode)
			}

			if tc.err != "" {
				var uR jsonapi.ErrorsPayload
				if err = json.NewDecoder(res.Body).Decode(&uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
				if uR.Errors[0].Detail != tc.err {
					t.Fatalf("expected error detail: %s, got: %s", uR.Errors[0].Detail, tc.err)
				}
			} else {
				var userList UserResponseList
				if _, err = jsonapi.UnmarshalManyPayload(res.Body, reflect.TypeOf(userList)); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)

				}
			}
		})

	}
}

func TestPatchUserHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	tt := []struct {
		id       string
		name     string
		username string
		password string
		status   int
		err      string
	}{
		{
			id:       "123",
			name:     "patch user information success",
			username: "admin",
			password: "AdminUser1*",
			status:   200,
		},
		{
			id:       "1000",
			name:     "patch user information - incorrect password",
			username: "admin",
			password: "test*",
			status:   400,
			err:      "password does not meet security requirements: password must be at least 8 characters long",
		},
	}

	rL, _ := models.NewRoleList([]string{models.AdminRole.String()})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 8)

	adminUser := &models.User{
		Model:    gorm.Model{ID: 123},
		Username: "admin",
		Version:  1,
		Roles:    rL,
	}
	adminUser.Password = string(hashedPassword)

	a := NewApp(s.DB, &Config{})

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			p := &UserRequest{
				ID:       tc.id,
				Username: tc.username,
				Password: tc.password,
				Roles:    rL.String(),
			}

			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, p); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("PATCH", "/user/123", requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			vars := map[string]string{
				"id": "123",
			}

			req = mux.SetURLVars(req, vars)

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()

			args := []driver.Value{int64(123)}
			queryArgs := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}
			twoArgs := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg()}
			nineArgs := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}

			if tc.status == http.StatusOK {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE "users" SET "updated_at"=$1,"username"=$2,"password"=$3,"version"=$4 WHERE "users"."deleted_at" IS NULL AND "id" = $5`)).
					WithArgs(queryArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "roles" ("created_at","updated_at","deleted_at","name","id") VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING RETURNING "id"`)).
					WithArgs(queryArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				s.mock.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO "user_roles" ("user_id","role_id") VALUES ($1,$2) ON CONFLICT DO NOTHING`)).
					WithArgs(twoArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
				s.mock.ExpectCommit()

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE "users" SET "updated_at"=$1 WHERE "users"."deleted_at" IS NULL AND "id" = $2`)).
					WithArgs(twoArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "roles" ("created_at","updated_at","deleted_at","name","id") VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING RETURNING "id"`)).
					WithArgs(queryArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				s.mock.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO "user_roles" ("user_id","role_id") VALUES ($1,$2) ON CONFLICT DO NOTHING`)).
					WithArgs(twoArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

				s.mock.ExpectCommit()

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectExec(regexp.QuoteMeta(
					`DELETE FROM "user_roles" WHERE "user_roles"."user_id" = $1 AND "user_roles"."role_id" <> $2`)).
					WithArgs(twoArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

				s.mock.ExpectCommit()

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "events" ("created_at","updated_at","deleted_at","username","activated","description","modified","authn_domain","severity") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WithArgs(nineArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()
			} else {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))
			}

			a.UserHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("expected status %d; got %d", tc.status, res.StatusCode)
			}

			if tc.err != "" {
				var uR jsonapi.ErrorsPayload
				if err = json.NewDecoder(res.Body).Decode(&uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
				if uR.Errors[0].Detail != tc.err {
					t.Fatalf("expected error detail: %s, got: %s", uR.Errors[0].Detail, tc.err)
				}
			} else {
				var uR UserResponse
				if err = jsonapi.UnmarshalPayload(res.Body, &uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	tt := []struct {
		id       string
		name     string
		username string
		password string
		status   int
		err      string
	}{
		{
			name:     "delete existing user",
			id:       "123",
			username: "admin",
			password: "AdminUser1*",
			status:   204,
		},
		{
			name:     "delete not existing user",
			id:       "",
			username: "admin",
			password: "TestUser1*",
			status:   404,
			err:      "user  not found in database",
		},
	}

	a := NewApp(s.DB, &Config{})

	rL, _ := models.NewRoleList([]string{models.AdminRole.String()})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	adminUser := &models.User{
		Model:    gorm.Model{ID: 1},
		Username: "admin",
		Version:  1,
		Roles:    rL,
	}
	adminUser.Password = string(hashedPassword)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			p := &UserRequest{
				ID:       tc.id,
				Username: tc.username,
				Password: tc.password,
			}

			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, p); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("DELETE", fmt.Sprintf("/user/%s", tc.id), requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			vars := map[string]string{
				"id": tc.id,
			}

			req = mux.SetURLVars(req, vars)

			var args []driver.Value

			if tc.id == "" {
				args = []driver.Value{int64(0)}
			} else {
				id, _ := strconv.Atoi(tc.id)
				args = []driver.Value{int64(id)}
			}

			nineArgs := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}

			// query
			if tc.status == 204 {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "Password", "Version"}).
					AddRow(tc.id, tc.name, tc.password, 1))

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectExec(regexp.QuoteMeta(
					`DELETE FROM "user_roles" WHERE "user_roles"."user_id" = $1`)).
					WithArgs(args...).WillReturnResult(sqlmock.NewResult(0, 1))

				s.mock.ExpectExec(regexp.QuoteMeta(
					`DELETE FROM "users" WHERE "users"."id" = $1`)).
					WithArgs(args...).WillReturnResult(sqlmock.NewResult(0, 1))

				s.mock.ExpectCommit()

				s.mock.ExpectBegin()
				s.EventRepo.DB.Begin()

				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "events" ("created_at","updated_at","deleted_at","username","activated","description","modified","authn_domain","severity") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WithArgs(nineArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()
			} else {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(tc.id).WillReturnRows(sqlmock.NewRows(nil)).WillReturnError(gorm.ErrRecordNotFound)
			}

			a.UserHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("expected status %d; got %d", tc.status, res.StatusCode)
			}

			if tc.err != "" {
				var uR jsonapi.ErrorsPayload
				if err = json.NewDecoder(res.Body).Decode(&uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
				if uR.Errors[0].Detail != tc.err {
					t.Fatalf("expected error detail: %s, got: %s", uR.Errors[0].Detail, tc.err)
				}
			}
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	tt := []struct {
		id       string
		name     string
		username string
		password string
		status   int
		err      string
	}{
		{
			name:     "get existing user",
			id:       "123",
			username: "admin",
			password: "AdminUser1*",
			status:   200,
		},
		{
			name:     "get NON existing user",
			id:       "345",
			username: "test",
			password: "AdminUser1*",
			status:   404,
			err:      "user 345 is not found",
		},
	}

	a := NewApp(s.DB, &Config{})

	rL, _ := models.NewRoleList([]string{models.AdminRole.String()})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	adminUser := &models.User{
		Model:    gorm.Model{ID: 123},
		Username: "admin",
		Version:  1,
		Roles:    rL,
	}
	adminUser.Password = string(hashedPassword)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			p := &UserRequest{
				ID:       tc.id,
				Username: tc.username,
				Password: tc.password,
			}

			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, p); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("GET", "/user/123", requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			vars := map[string]string{
				"id": tc.id,
			}

			req = mux.SetURLVars(req, vars)

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()

			var args []driver.Value

			if tc.id == "" {
				args = []driver.Value{int64(0)}
			} else {
				id, _ := strconv.Atoi(tc.id)
				args = []driver.Value{int64(id)}
			}

			if tc.status == http.StatusOK {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))
			} else {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows(nil)).WillReturnError(gorm.ErrRecordNotFound)
			}
			a.UserHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("expected status %d; got %d", tc.status, res.StatusCode)
			}

			if tc.err != "" {
				var uR jsonapi.ErrorsPayload
				if err = json.NewDecoder(res.Body).Decode(&uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
				if uR.Errors[0].Detail != tc.err {
					t.Fatalf("expected error detail: %s, got: %s", uR.Errors[0].Detail, tc.err)
				}
			} else {
				var uR UserResponse
				if err = jsonapi.UnmarshalPayload(res.Body, &uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)

				}
			}
		})

	}
}
