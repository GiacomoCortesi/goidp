package models

import (
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBConnect(dsn string, logger logger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err.Error())
	}

	if err := db.AutoMigrate(&User{}, &Role{}, &Event{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %s", err.Error())
	}

	return db, nil
}

// DBError wraps all the DB related errors
// this shall be normally mapped as Internal Server Error HTTP errors
type DBError struct {
	error string
}

func (e *DBError) Error() string {
	return e.error
}

// NotFoundError wraps the not found errors
// this shall be normally mapped as Not Found HTTP errors
type NotFoundError struct {
	error string
}

func (e *NotFoundError) Error() string {
	return e.error
}

// UserError wraps all the database errors caused by mistaken user input
// e.g. invalid password
// this shall be normally mapped as Bad Request HTTP errors
type UserError struct {
	error string
}

func (e *UserError) Error() string {
	return e.error
}
