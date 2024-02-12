package controllers

import (
	"net/http"
)

type GetSystemResponse struct {
	Data []*SystemResponseData
}

type SystemResponseData struct {
	ID           string `jsonapi:"primary,app"`
	AppBuild     string `jsonapi:"attr,app_build"`
	AppName      string `jsonapi:"attr,app_name"`
	ChartVersion string `jsonapi:"attr,chart_version"`
	AppVersion   string `jsonapi:"attr,app_version"`
	ApiVersion   string `jsonapi:"attr,api_version"`
}

func (a *App) SystemHandler(w http.ResponseWriter, r *http.Request) {
	var getSystemResponse *GetSystemResponse
	var systemResponseData []*SystemResponseData

	systemResponseData = append(systemResponseData, &SystemResponseData{
		ID:           "idp",
		AppBuild:     AppBuild,
		AppName:      AppName,
		ChartVersion: ChartVersion,
		AppVersion:   AppVersion,
		ApiVersion:   ApiVersion,
	})

	getSystemResponse = &GetSystemResponse{
		Data: systemResponseData,
	}

	jsonapiSuccess(w, getSystemResponse.Data, http.StatusOK)
}
