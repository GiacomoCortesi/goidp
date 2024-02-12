package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/jsonapi"
)

// jsonapiMiddleware is a middleware function that makes sure the client provides the proper media type in the request
func (a *App) jsonapiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Servers MUST respond with a 415 Unsupported Media Type status code if a request specifies the header
		// Content-Type: application/vnd.api+json with any media type parameters.
		if r.Header.Get(headerContentType) != "" {
			m, _ := regexp.MatchString(regexp.QuoteMeta(jsonapi.MediaType)+`\s*;`, r.Header.Get(headerContentType))
			if m {
				jsonapiError(w, http.StatusUnsupportedMediaType, "unsupported Media Type")
				return
			}
		}
		// Servers MUST respond with a 406 Not Acceptable status code if a request’s Accept header contains the JSON:API
		// media type and all instances of that media type are modified with media type parameters.
		if r.Header.Get(headerAccept) != "" && strings.Contains(r.Header.Get(headerAccept), jsonapi.MediaType) {
			m, _ := regexp.MatchString(regexp.QuoteMeta(jsonapi.MediaType)+`\s*,|`+regexp.QuoteMeta(jsonapi.MediaType)+`\s*$`, r.Header.Get(headerAccept))
			if !m {
				jsonapiError(w, http.StatusNotAcceptable, "Not Acceptable")
				return
			}
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// jsonapiSuccess formats and return a successful response in the jsonapi format
func jsonapiSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set(headerContentType, jsonapi.MediaType)
	w.WriteHeader(statusCode)
	if err := jsonapi.MarshalPayload(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// jsonapiNoContentSuccess just writes the headers for No Content http response
func jsonapiNoContentSuccess(w http.ResponseWriter) {
	w.Header().Set(headerContentType, jsonapi.MediaType)
	w.WriteHeader(http.StatusNoContent)
}

// jsonapiSuccessMetaOnly is similar to jsonapiSuccess but the jsonapi response only
// contains Meta information. Data field is set to null.
// NOTE: You cannot avoid having the Data field serialized since it does not have omitempty
// tag. However, as per openapi spec, it should be allowed to just have meta as a
// top level key (debatable whether this makes sense or not)
func jsonapiSuccessMetaOnly(w http.ResponseWriter, data jsonapi.Metable, statusCode int) {
	w.Header().Set(headerContentType, jsonapi.MediaType)
	w.WriteHeader(statusCode)

	var p jsonapi.OnePayload
	if metableModel, ok := data.(jsonapi.Metable); ok {
		p.Meta = metableModel.JSONAPIMeta()
	}
	if err := json.NewEncoder(w).Encode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// jsonapiSuccessWithMeta is similar to jsonapiSuccess but the jsonapi response
// contains also outer Meta information.
func jsonapiSuccessWithMeta(w http.ResponseWriter, meta jsonapi.Metable, data interface{}, statusCode int) {
	w.Header().Set(headerContentType, jsonapi.MediaType)
	w.WriteHeader(statusCode)
	p, err := jsonapi.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	switch p.(type) {
	case *jsonapi.ManyPayload:
		p.(*jsonapi.ManyPayload).Meta = meta.JSONAPIMeta()
	case *jsonapi.OnePayload:
		p.(*jsonapi.OnePayload).Meta = meta.JSONAPIMeta()
	}

	if err := json.NewEncoder(w).Encode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// jsonapiError formats and return an error response in the jsonapi format
func jsonapiError(w http.ResponseWriter, code int, message string) {
	w.Header().Set(headerContentType, jsonapi.MediaType)
	w.WriteHeader(code)
	// NOTE: ErrorObject as defined by jsonapi module doesn't include source field. Do we need it for real?
	eO := &jsonapi.ErrorObject{
		Detail: message,
		Status: strconv.Itoa(code),
		Code:   strconv.Itoa(code),
	}
	if err := jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{eO}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
