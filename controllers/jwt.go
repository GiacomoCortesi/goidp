package controllers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/goidp/models"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type customClaims struct {
	Roles []string `json:"roles"`
	Azt   string   `json:"azt"`
	jwt.StandardClaims
}

func (a *App) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := parseAuthHeader(r)
		claims, err := getClaimsFromAccessToken(t, a.config.Secret, a.config.VerifyKey)
		if err != nil {
			log.WithFields(log.Fields{
				"claims": claims,
				"error":  err,
			}).Info("unauthorized request")
			jsonapiError(w, http.StatusUnauthorized, "unauthorized request")
			return
		} else {
			if !stringInSliceCaseInsensitive(claims.Roles, models.AdminRole.String()) {
				// non-admin users cannot POST/DELETE
				if r.Method == http.MethodPost || r.Method == http.MethodDelete {
					jsonapiError(w, http.StatusForbidden, "forbidden request")
					return
				}
				params := mux.Vars(r)
				id, ok := params["id"]
				// non-admin users can only PATCH themselves
				if ok && r.Method == http.MethodPatch && id != claims.Subject {
					jsonapiError(w, http.StatusForbidden, "forbidden request")
					return
				}
			}
			next.ServeHTTP(w, r)
		}
	})
}

func newStandardClaims(user *models.User, issuer string, expire time.Duration) jwt.StandardClaims {
	return jwt.StandardClaims{
		ExpiresAt: time.Now().Add(expire).Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   user.Username,
		Id:        uuid.New().String(),
		Issuer:    issuer,
		NotBefore: time.Now().Unix(),
	}
}

func newCustomClaims(user *models.User, domain string, expire time.Duration) customClaims {
	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}

	return customClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expire).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   user.Username,
			Id:        uuid.New().String(),
			Issuer:    "idp",
			NotBefore: time.Now().Unix(),
		},
		Roles: roles,
		Azt:   domain,
	}
}

func generateToken(claims jwt.Claims, secret string, signKey *rsa.PrivateKey) (string, error) {
	var signedToken string
	var err error
	if secret == "" {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err = token.SignedString(signKey)
	} else {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err = token.SignedString([]byte(secret))
	}
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func getKeyFunc(secret string, verifyKey *rsa.PublicKey) jwt.Keyfunc {
	var keyFunc jwt.Keyfunc
	if secret == "" {
		keyFunc = func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return verifyKey, nil
		}
	} else {
		keyFunc = func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		}
	}
	return keyFunc
}

func getClaimsFromRenewToken(t string, secret string, verifyKey *rsa.PublicKey) (*jwt.StandardClaims, error) {
	keyFunc := getKeyFunc(secret, verifyKey)
	token, err := jwt.ParseWithClaims(
		t,
		&jwt.StandardClaims{},
		keyFunc,
	)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"token": token,
	}).Debug("jwt token")
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, errors.New("error decoding claims")
	}
	log.WithFields(log.Fields{
		"expire_time": c.ExpiresAt,
		"issued_at":   c.IssuedAt,
		"subject":     c.Subject,
		"issuer":      c.Issuer,
	}).Info("jwt token claims")
	return c, nil
}

func getClaimsFromM2MToken(t string, verifyKey *rsa.PublicKey) (*customClaims, error) {
	token, err := jwt.ParseWithClaims(
		t,
		&customClaims{},
		func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("error jwt method not allowed")
			}
			return verifyKey, nil
		})
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"token": token,
	}).Debug("jwt token")
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := token.Claims.(*customClaims)
	if !ok {
		return nil, errors.New("error decoding claims")
	}
	log.WithFields(log.Fields{
		"expire_time": c.ExpiresAt,
		"issued_at":   c.IssuedAt,
		"subject":     c.Subject,
		"issuer":      c.Issuer,
		"role_list":   c.Roles,
	}).Info("jwt token claims")

	return c, nil
}

func getClaimsFromAccessToken(t string, secret string, verifyKey *rsa.PublicKey) (*customClaims, error) {
	keyFunc := getKeyFunc(secret, verifyKey)
	token, err := jwt.ParseWithClaims(
		t,
		&customClaims{},
		keyFunc,
	)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"token": token,
	}).Debug("jwt token")
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := token.Claims.(*customClaims)
	if !ok {
		return nil, errors.New("error decoding claims")
	}
	log.WithFields(log.Fields{
		"expire_time": c.ExpiresAt,
		"issued_at":   c.IssuedAt,
		"subject":     c.Subject,
		"issuer":      c.Issuer,
		"role_list":   c.Roles,
	}).Info("jwt token claims")

	return c, nil
}

func parseAuthHeader(r *http.Request) string {
	t := r.Header.Get(headerAuthorization)
	sT := strings.SplitAfter(t, "Bearer")
	var rawT string
	switch len(sT) {
	case 1:
		// we also accept requests without Bearer keyword
		rawT = sT[0]
	case 2:
		rawT = sT[1]
	}
	return strings.TrimSpace(rawT)
}

func authorizeM2MRequest(t string, pubKeys []*rsa.PublicKey) (*customClaims, error) {
	if pubKeys == nil {
		return nil, fmt.Errorf("no trusted public keys configured")
	}
	if t == "" {
		return nil, fmt.Errorf("empty token")
	}

	for _, pKey := range pubKeys {
		c, err := getClaimsFromM2MToken(t, pKey)
		if err != nil {
			log.WithError(err).Info("unable to extract claims from token")
			continue
		}
		return c, nil
	}

	return nil, fmt.Errorf("no issuer found")
}
