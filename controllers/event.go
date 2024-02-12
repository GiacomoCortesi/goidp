package controllers

import (
	"github.com/goidp/models"
	"net/http"
	"strconv"
	"time"

	"github.com/google/jsonapi"
)

const (
	defaultPageSize   int = 25
	defaultPageNumber int = 1
)

// EventsHandler handles an event request
func (a *App) EventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.GetEventsHandler(w, r)
	}
}

type GetEventsJSONAPIResponse struct {
	ID          uint   `jsonapi:"primary,event,omitempty" json:"id,omitempty"`
	Activated   string `jsonapi:"attr,activated" json:"activated,omitempty"`
	AuthnDomain string `jsonapi:"attr,authn_domain" json:"authn_domain,omitempty"`
	Description string `jsonapi:"attr,description" json:"description,omitempty"`
	Modified    string `jsonapi:"attr,modified" json:"modified,omitempty"`
	Severity    string `jsonapi:"attr,severity" json:"severity,omitempty"`
	Username    string `jsonapi:"attr,username" json:"username,omitempty"`
}

type GetEventsJSONAPIResponseMeta struct {
	eS         *EventSummary
	totalPages int
}

type SeverityCounts struct {
	Cleared       int `json:"cleared"`
	Critical      int `json:"critical"`
	Indeterminate int `json:"indeterminate"`
	Major         int `json:"major"`
	Minor         int `json:"minor"`
	Warning       int `json:"warning"`
}

type EventSummary struct {
	SeverityCounts *SeverityCounts `json:"severity_counts"`
}

func (g GetEventsJSONAPIResponseMeta) JSONAPIMeta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"summary":     g.eS,
		"total_pages": g.totalPages,
	}
}

// GetEventsHandler is the function that handles a GET request for an event
// and performs a query to DB in order to retrieve the event if present
func (a *App) GetEventsHandler(w http.ResponseWriter, r *http.Request) {
	var getEventsJSONAPIResponses []*GetEventsJSONAPIResponse

	queryValues := r.URL.Query()

	pageNumber, err := strconv.Atoi(queryValues.Get("page[number]"))
	if err != nil {
		pageNumber = defaultPageNumber
	}
	pageSize, err := strconv.Atoi(queryValues.Get("page[size]"))
	if err != nil {
		pageSize = defaultPageSize
	}

	// summary default to false if no summary query params passed
	summary, _ := strconv.ParseBool(queryValues.Get("summary"))

	events := a.Events.GetEvents(pageNumber, pageSize)

	for _, e := range events {
		getEventsJSONAPIResponses = append(getEventsJSONAPIResponses, &GetEventsJSONAPIResponse{
			ID:          e.ID,
			Activated:   e.Activated.Format(time.RFC3339),
			AuthnDomain: e.AuthnDomain,
			Description: e.Description,
			Modified:    e.Modified.Format(time.RFC3339),
			Severity:    e.Severity.String(),
			Username:    e.Username,
		},
		)
	}

	if !summary {
		jsonapiSuccess(w, getEventsJSONAPIResponses, http.StatusOK)
		return
	}
	var sC SeverityCounts
	sC.Cleared = a.Events.GetEventsCount(models.EventSeverityCleared)
	sC.Indeterminate = a.Events.GetEventsCount(models.EventSeverityIndeterminate)
	sC.Warning = a.Events.GetEventsCount(models.EventSeverityWarning)
	sC.Minor = a.Events.GetEventsCount(models.EventSeverityMinor)
	sC.Major = a.Events.GetEventsCount(models.EventSeverityMajor)
	sC.Critical = a.Events.GetEventsCount(models.EventSeverityCritical)
	jsonapiSuccessWithMeta(w, GetEventsJSONAPIResponseMeta{
		eS:         &EventSummary{SeverityCounts: &sC},
		totalPages: 1 + ((sC.Major + sC.Minor + sC.Critical + sC.Cleared + sC.Warning + sC.Indeterminate) / pageSize),
	}, getEventsJSONAPIResponses, http.StatusOK)
}
