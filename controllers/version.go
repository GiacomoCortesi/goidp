package controllers

import (
	"net/http"
)

const (
	ApiVersion = "v1.0"
)

type GetVersionsResponse struct {
	ID         int    `jsonapi:"primary,versions"`
	Version    string `jsonapi:"attr,version"`
	Deprecated bool   `jsonapi:"attr,deprecated"`
}

func (a *App) GetVersions(w http.ResponseWriter, r *http.Request) {
	v := GetVersionsResponse{
		ID:         1,
		Version:    ApiVersion,
		Deprecated: false,
	}
	jsonapiSuccess(w, []*GetVersionsResponse{&v}, http.StatusOK)
}
