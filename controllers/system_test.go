package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemHandler(t *testing.T) {
	s, err := SetupSuite()
	assert.Nil(t, err)
	app := NewApp(s.DB, &Config{})

	err = os.Setenv("APP_BUILD", "build")
	assert.Nil(t, err)

	err = os.Setenv("APP_NAME", "idp")
	assert.Nil(t, err)

	err = os.Setenv("APP_RELEASE", "1.0.0")
	assert.Nil(t, err)

	err = os.Setenv("APP_VERSION", "1")
	assert.Nil(t, err)

	request, _ := http.NewRequest(http.MethodGet, "/system", http.NoBody)
	response := httptest.NewRecorder()

	app.SystemHandler(response, request)

	assert.Equal(t, response.Code, 200)
}
