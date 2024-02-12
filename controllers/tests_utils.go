package controllers

import (
	"github.com/goidp/models"
	"os"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Suite is the struct that holds unit tests needed data
type Suite struct {
	DB        *gorm.DB
	mock      sqlmock.Sqlmock
	EventRepo *models.EventRepo
	UserRepo  *models.UserRepo
	RoleRepo  *models.RoleRepo
}

// SetupSuite sets up Suite struct
// this method initializes mocked connection to db and sets necessary environment variables
func SetupSuite() (*Suite, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := setEnvironmentVariables(); err != nil {
		return nil, err
	}

	return &Suite{
		mock:      mock,
		DB:        gdb,
		EventRepo: &models.EventRepo{DB: gdb},
	}, nil
}

// setEnvironmentVariables is the utility function to set needed env variables
func setEnvironmentVariables() error {
	err := os.Setenv("APP_WRITE_TIMEOUT", "10")
	if err != nil {
		return err
	}

	err = os.Setenv("APP_READ_TIMEOUT", "15")
	if err != nil {
		return err
	}

	err = os.Setenv("APP_IDLE_TIMEOUT", "20")
	if err != nil {
		return err
	}
	return nil
}
