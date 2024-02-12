package controllers

import (
	"database/sql/driver"
	"github.com/goidp/models"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/jsonapi"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetEventsHandler(t *testing.T) {
	e := &models.Event{
		Model:       gorm.Model{ID: 123},
		Username:    "aorsi",
		Activated:   time.Now(),
		Description: "Login successful from IP ::ffff:10.150.100.117",
		Modified:    time.Now(),
		AuthnDomain: models.ExternalDomain,
		Severity:    models.EventSeverityWarning,
	}
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf(err.Error())
	}

	a := NewApp(s.DB, &Config{})

	req, err := http.NewRequest("GET", "/event?page[number]=1&page[size]=25&summary=true", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rec := httptest.NewRecorder()

	s.mock.ExpectBegin()
	s.EventRepo.DB.Begin()

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL ORDER BY id desc LIMIT 25`)).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"Username", "Activated", "Description",
			"Modified", "AuthnDomain", "Severity"}).
			AddRow(e.Username, e.Activated, e.Description,
				e.Modified, e.AuthnDomain, e.Severity))
	for i := 1; i <= 6; i++ {
		s.mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "events" WHERE severity = $1 AND "events"."deleted_at" IS NULL`,
		)).WithArgs(i).WillReturnRows(sqlmock.NewRows([]string{""}))
	}
	a.GetEventsHandler(rec, req)

	if err := s.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	res := rec.Result()

	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Errorf("expected status %d; got %d", 200, res.StatusCode)
	}

	_, err = jsonapi.UnmarshalManyPayload(res.Body, reflect.TypeOf([]GetEventsJSONAPIResponse{}))
	if err != nil {
		t.Fatalf("could not unmarshal response %s", err)
	}
}

func TestStoreEvent(t *testing.T) {
	e := &models.Event{
		Model:       gorm.Model{ID: 123},
		Username:    "aorsi",
		Activated:   time.Now(),
		Description: "Login successful from IP ::ffff:10.150.100.117",
		Modified:    time.Now(),
		AuthnDomain: models.ExternalDomain,
		Severity:    models.EventSeverityWarning,
	}

	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	s.mock.ExpectBegin()
	s.EventRepo.DB.Begin()

	args := []driver.Value{int64(123)}

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND "events"."id" = $1 ORDER BY "events"."id" LIMIT 1`)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows(nil))

	insertArgs := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), e.Username, e.Activated, e.Description, e.Modified, e.AuthnDomain, e.Severity, e.ID}

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "events" ("created_at","updated_at","deleted_at","username","activated","description","modified","authn_domain","severity","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING "id"`)).
		WithArgs(insertArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	s.mock.ExpectCommit()

	err = s.EventRepo.Create(e)

	assert.Nil(t, err)

	if err := s.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
