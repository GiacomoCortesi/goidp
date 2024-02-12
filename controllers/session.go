package controllers

import (
	"fmt"
	"github.com/goidp/models"
	"net/http"
	"strconv"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	headerAccept        = "Accept"
)

// SessionHandler is the function which verifies what action to take based on the request (if POST or DELETE)
func (a *App) SessionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.CreateSessionHandler(w, r)
	case http.MethodDelete:
		a.DeleteSessionHandler(w, r)
	}
}

type CreateSessionHandlerResponse struct {
	AccessToken string
	RenewToken  string
}

func (sessionHandlerResponse CreateSessionHandlerResponse) JSONAPIMeta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"access_token": sessionHandlerResponse.AccessToken,
		"renew_token":  sessionHandlerResponse.RenewToken,
	}
}

type CreateSessionHandlerRequest struct {
	ID          string `jsonapi:"primary,session,omitempty"`
	Username    string `jsonapi:"attr,username,omitempty"`
	Password    string `jsonapi:"attr,password,omitempty"`
	AccessToken string `jsonapi:"attr,access_token,omitempty"` // AccessToken is provided in case of m2m authentication
}

func (a *App) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	var claims customClaims
	var user = &models.User{}
	var domain string

	params := mux.Vars(r)
	// validate default to false if no validate query params passed
	validate, _ := strconv.ParseBool(params["validate"])

	var requestBody CreateSessionHandlerRequest
	if err := jsonapi.UnmarshalPayload(r.Body, &requestBody); err != nil {
		jsonapiError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err.Error()))
		return
	}

	// we accept a JWT both if it passed as Auth header and in request body
	var t string
	if requestBody.AccessToken != "" {
		t = requestBody.AccessToken
	} else {
		t = parseAuthHeader(r)
	}

	ip, _ := getIP(r)
	if t != "" {
		// authentication with token
		decodedClaims, err := authorizeM2MRequest(t, a.config.TrustedPublicKeys)
		if err != nil {
			log.WithError(err).Warnf("authorization failure")
			jsonapiError(w, http.StatusUnauthorized, "supplied token is not valid")

			err := a.Events.CreateUnsuccessfulLoginEvent("unknown", models.ExternalDomain, ip)
			if err != nil {
				log.WithError(err).Warnf("failed to store login attempt")
			}
			return
		}
		user.Username = decodedClaims.Subject
		roles, err := models.NewRoleList(decodedClaims.Roles)
		if err != nil {
			log.WithError(err).Warnf("unable to convert roles")
		} else {
			user.Roles = roles
		}
		domain = models.ExternalDomain
		a.extUsers[user.Username] = roles
	} else {
		// authentication with credentials
		var ok bool
		user, ok = a.Users.GetAndValidateUser(requestBody.Username, requestBody.Password)
		if !ok {
			err := a.Events.CreateUnsuccessfulLoginEvent(requestBody.Username, models.InternalDomain, ip)
			if err != nil {
				log.WithError(err).Warnf("failed to store login attempt")
			}
			jsonapiError(w, http.StatusUnauthorized, "user not allowed")
			return
		}
		domain = models.InternalDomain
	}

	err := a.Events.CreateSuccessfulLoginEvent(user.Username, domain, ip)
	if err != nil {
		log.WithError(err).Warnf("failed to store login attempt")
	}

	if validate {
		// if validate set to true, returned jwt token expires immediately, so that it cannot be used for
		// subsequent requests
		claims = newCustomClaims(user, domain, 0)
	} else {
		// if validate set to false, we create the token with default expire time
		claims = newCustomClaims(user, domain, a.config.AccessTokenExpireTime)
	}

	signedAccessToken, err := generateToken(claims, a.config.Secret, a.config.SignKey)
	if err != nil {
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		err = a.Events.CreateJWTEvent(user.Username, models.InternalDomain)
		if err != nil {
			log.WithError(err).Warnf("failed to store JWT event")
		}
		return
	}
	w.Header().Set(headerAuthorization, fmt.Sprintf("Bearer %v", signedAccessToken))

	var responseBody CreateSessionHandlerResponse
	responseBody.AccessToken = signedAccessToken
	if a.config.RenewTokenExpireTime == 0 {
		// no renew token functionality configured
		jsonapiSuccessMetaOnly(w, &responseBody, http.StatusOK)
	} else {
		rC := newStandardClaims(user, domain, a.config.RenewTokenExpireTime)
		signedRenewToken, err := generateToken(rC, a.config.Secret, a.config.SignKey)
		if err != nil {
			jsonapiError(w, http.StatusInternalServerError, err.Error())
			return
		}
		responseBody.RenewToken = signedRenewToken
		jsonapiSuccessMetaOnly(w, &responseBody, http.StatusOK)
	}
}

func (a *App) DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: How do we handle session backend side?
	// so far no jwt information is stored into db, fully stateless
}

type RenewTokenHandlerRequest struct {
	UserID     string `jsonapi:"primary,renew"`
	RenewToken string `jsonapi:"attr,renew_token"`
}

func (a *App) RenewTokenHandler(w http.ResponseWriter, r *http.Request) {
	var renewTokenHandlerRequest RenewTokenHandlerRequest
	var err error
	if err = jsonapi.UnmarshalPayload(r.Body, &renewTokenHandlerRequest); err != nil {
		jsonapiError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err.Error()))
		return
	}

	var user *models.User

	if roles, ok := a.extUsers[renewTokenHandlerRequest.UserID]; ok {
		user = &models.User{
			Username: renewTokenHandlerRequest.UserID,
			Roles:    roles,
		}
	} else {
		user, err = a.Users.GetUserByNameOrID(renewTokenHandlerRequest.UserID)
		switch err.(type) {
		case *models.NotFoundError:
			jsonapiError(w, http.StatusUnauthorized, "user revoked")
			return
		case *models.DBError:
			jsonapiError(w, http.StatusInternalServerError, "internal error, retry later")
			return
		}
	}

	claims, err := getClaimsFromRenewToken(renewTokenHandlerRequest.RenewToken, a.config.Secret, a.config.VerifyKey)
	if err != nil {
		log.WithFields(log.Fields{
			"token": renewTokenHandlerRequest.RenewToken,
			"error": err,
		}).Info("invalid renew token")
		jsonapiError(w, http.StatusUnauthorized, "invalid renew token")
		return
	}
	accessClaims := newCustomClaims(user, claims.Issuer, a.config.RenewTokenExpireTime)
	signedAccessToken, err := generateToken(accessClaims, a.config.Secret, a.config.SignKey)
	if err != nil {
		jsonapiError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set(headerAuthorization, fmt.Sprintf("Bearer %v", signedAccessToken))

	// this is an implementation choice, we can either:
	// - generate a new renew token at every renew request
	// - return the same renew token
	// for now we use same renew token
	//refreshClaims := newStandardClaims(user)
	//signedRefreshToken, err := generateToken(refreshClaims)
	//if err != nil {
	//	jsonapiError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}
	var responseBody CreateSessionHandlerResponse
	responseBody.AccessToken = signedAccessToken
	responseBody.RenewToken = renewTokenHandlerRequest.RenewToken

	jsonapiSuccessMetaOnly(w, &responseBody, http.StatusOK)
}
