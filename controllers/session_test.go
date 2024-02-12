package controllers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"github.com/goidp/models"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt"
	"github.com/google/jsonapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateSessionHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	tt := []struct {
		name     string
		username string
		password string
		status   int
		err      string
	}{
		{
			name:     "empty credentials",
			username: "",
			password: "",
			status:   401,
			err:      "user not allowed",
		},
		{
			name:     "wrong credentials",
			username: "ammacca",
			password: "banane",
			status:   401,
			err:      "user not allowed",
		},
		{
			name:     "correct credentials",
			username: "admin",
			password: "admin",
			status:   200,
		},
	}

	rL, _ := models.NewRoleList([]string{models.AdminRole.String()})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	adminUser := &models.User{
		Username: "admin",
		Version:  0,
		Roles:    rL,
	}
	adminUser.Password = string(hashedPassword)

	a := NewApp(s.DB, &Config{})
	pKey, _ := ReadPrivateKey("-----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAKCAQEAyqUnBHSpEaaMEqZ2NvTZtK1bC7IPCQKWN8oi0fWZIxqzz4zE\nSTYyLdLV/q1cHYVb7hx53QoUNvUUm1SNAqMs4vWI+Dj+hbcnMUusKPB9PQK+WP06\nSTjQCqW6rGzpzftK0tW7leOy7tLSFVMcEvzkKM+04zRYY8+oFh/8rzdHypTIdYWp\nSyW9zHpnhA16yP7TNqZFlWosGLRSG6mfQtK0Mb8KBqsjHah6jzAVltRkfFnPb62x\nUcQfVeaf17wXJN6XFtH5GSEi0snrdiMvVCtnco6NBFzjNeVnh4BJa5+TLvdgE70z\n2QGu8nAV7BJrXBFNFpYTeKumqJ/owgzM2gGo2QIDAQABAoIBAQCmZl0WrJEUPFVz\nDxutXvvSADPt86Wi+WvOnf5fuDOqfse+G1Im6AjmVeWA/mvQlex6JwnudtNImZD1\nR8WOr90w9PwnD+34cQAO25uf9nJwges5+Z49+BflVldmNPz8Nmgnnngtyc7pi1YV\nSqyX7u+Pj5dypk4aj67vlA6S9mrOLk4G8IsJwZt5Hdsp/Erze6wHM7snuZL+acM4\n2FQ7syZ4epvI782y2sRuaiWZMY96CH7PElqWUiz79/asSa1TZaPQ7sDD3AtVqMJm\n5jGeeKwdJJNe+KSLk9gVFnBvfUm38S9UAp/NgjZPFFurGN3ElhZHayK4Q4UL3P/a\n0X2jd7MBAoGBAOn9zgeDNZ3YY0RdAJhyxPGjdER41hb6Gs4Tum/dzRokfygVpCba\nihp42bZqGF1KPXIYZBEky1wYrjOskiiFkICrzJTSMtbHak3+u8hXyYFG0TOZz79y\nvwB4fJp+2G17C7S6uD8a91DZZpQ8dpNWDpvbh1WbZnx3CjbcbJ870YFhAoGBAN20\nkkZpQW5XXQpoPRUuXSljSJk8Gsi8X2yuOTWDFGyDr2VUDAeTd7Jm9PirINOZHD3e\n1rAQ01GMzRdYPC1czULerlRwH3lxN2r3ovU7f1Zu3iCpgrqVSRoeX1R0tisdf62z\nl/V6o1EUxsMEV+Uboccew/hJ0xvNRKLmHmil8MJ5AoGADLp2s6fqibyUocphVumf\nVvmqQHNGSheuz5j5Ik6xcoObuyV6OXbX3lrGlQquapy4PPWgs+IJgegByePQS44A\nb09pIItSoqZUXQvHUT2dQ4ADr0flqidmxnLHbGwL/+CaoWkqzpv76hT5ZITpelhL\nESVe9kQuzgR3tMZGzl6lpeECgYEAt7f+zuJCGlHDA/DFTVwST026x2CLQXT4DnOB\nbNqmfhXRrsIrBcwqEGhI8Be/KBlk0dBrT5Nhyd5HxeSUWXLhlVw6UjZnnpc3OSjk\nnRsktldBMwfFESDMZxxsGuxsWOYk+6grcHykAXiaDNj4jR6MvRi9hG6Ixi0fh23y\nHP4FuOECgYEA3+DT0VORSxntxpYviM6fulJFDhNG8QnqsVlEVCBhjQ2F02yxwMN9\nlHTCzKc9BZMlVhRei0ZqluCVSyspqYmV3ufa0CWlGhC4JwnHmI6e8bXmu/SZBkxP\nwwU/2aP8w4viG3s4S+UpE7Zkdre78mS/SUWwb2ShlJFjG7CcBUWrs/E=\n-----END RSA PRIVATE KEY-----\n")

	a.config.SignKey = pKey

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			p := &CreateSessionHandlerRequest{
				Username: tc.username,
				Password: tc.password,
			}
			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, p); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("POST", "/session", requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()

			s.mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
				WithArgs(tc.username).
				WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))

			insertArgs := []driver.Value{sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()
			if tc.status == http.StatusUnauthorized {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "events" ("created_at","updated_at","deleted_at","username","activated","description","modified","authn_domain","severity") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WithArgs(insertArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()
			} else if tc.status == http.StatusOK {
				s.mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "events" ("created_at","updated_at","deleted_at","username","activated","description","modified","authn_domain","severity") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WithArgs(insertArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()
			}

			a.SessionHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatalf("expected status %d; got %d", tc.status, res.StatusCode)
			}
			if tc.err != "" {
				var uR jsonapi.ErrorsPayload
				if err = json.NewDecoder(res.Body).Decode(&uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
				if uR.Errors[0].Detail != tc.err {
					t.Fatalf("expected error detail: %s, got: %s", uR.Errors[0].Detail, tc.err)
				}
			} else {
				var uR CreateSessionHandlerResponse
				if err = jsonapi.UnmarshalPayload(res.Body, &uR); err != nil {
					t.Fatalf("could not unmarshal response: %s", err)
				}
			}
		})
	}
}

func TestRenewTokenHandler(t *testing.T) {
	s, err := SetupSuite()
	if err != nil {
		t.Fatalf("error setting up test suite: %s", err.Error())
	}

	a := NewApp(s.DB, &Config{})

	// we need keys to encode/decode tokens
	pKey, _ := ReadPrivateKey("-----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAKCAQEAyqUnBHSpEaaMEqZ2NvTZtK1bC7IPCQKWN8oi0fWZIxqzz4zE\nSTYyLdLV/q1cHYVb7hx53QoUNvUUm1SNAqMs4vWI+Dj+hbcnMUusKPB9PQK+WP06\nSTjQCqW6rGzpzftK0tW7leOy7tLSFVMcEvzkKM+04zRYY8+oFh/8rzdHypTIdYWp\nSyW9zHpnhA16yP7TNqZFlWosGLRSG6mfQtK0Mb8KBqsjHah6jzAVltRkfFnPb62x\nUcQfVeaf17wXJN6XFtH5GSEi0snrdiMvVCtnco6NBFzjNeVnh4BJa5+TLvdgE70z\n2QGu8nAV7BJrXBFNFpYTeKumqJ/owgzM2gGo2QIDAQABAoIBAQCmZl0WrJEUPFVz\nDxutXvvSADPt86Wi+WvOnf5fuDOqfse+G1Im6AjmVeWA/mvQlex6JwnudtNImZD1\nR8WOr90w9PwnD+34cQAO25uf9nJwges5+Z49+BflVldmNPz8Nmgnnngtyc7pi1YV\nSqyX7u+Pj5dypk4aj67vlA6S9mrOLk4G8IsJwZt5Hdsp/Erze6wHM7snuZL+acM4\n2FQ7syZ4epvI782y2sRuaiWZMY96CH7PElqWUiz79/asSa1TZaPQ7sDD3AtVqMJm\n5jGeeKwdJJNe+KSLk9gVFnBvfUm38S9UAp/NgjZPFFurGN3ElhZHayK4Q4UL3P/a\n0X2jd7MBAoGBAOn9zgeDNZ3YY0RdAJhyxPGjdER41hb6Gs4Tum/dzRokfygVpCba\nihp42bZqGF1KPXIYZBEky1wYrjOskiiFkICrzJTSMtbHak3+u8hXyYFG0TOZz79y\nvwB4fJp+2G17C7S6uD8a91DZZpQ8dpNWDpvbh1WbZnx3CjbcbJ870YFhAoGBAN20\nkkZpQW5XXQpoPRUuXSljSJk8Gsi8X2yuOTWDFGyDr2VUDAeTd7Jm9PirINOZHD3e\n1rAQ01GMzRdYPC1czULerlRwH3lxN2r3ovU7f1Zu3iCpgrqVSRoeX1R0tisdf62z\nl/V6o1EUxsMEV+Uboccew/hJ0xvNRKLmHmil8MJ5AoGADLp2s6fqibyUocphVumf\nVvmqQHNGSheuz5j5Ik6xcoObuyV6OXbX3lrGlQquapy4PPWgs+IJgegByePQS44A\nb09pIItSoqZUXQvHUT2dQ4ADr0flqidmxnLHbGwL/+CaoWkqzpv76hT5ZITpelhL\nESVe9kQuzgR3tMZGzl6lpeECgYEAt7f+zuJCGlHDA/DFTVwST026x2CLQXT4DnOB\nbNqmfhXRrsIrBcwqEGhI8Be/KBlk0dBrT5Nhyd5HxeSUWXLhlVw6UjZnnpc3OSjk\nnRsktldBMwfFESDMZxxsGuxsWOYk+6grcHykAXiaDNj4jR6MvRi9hG6Ixi0fh23y\nHP4FuOECgYEA3+DT0VORSxntxpYviM6fulJFDhNG8QnqsVlEVCBhjQ2F02yxwMN9\nlHTCzKc9BZMlVhRei0ZqluCVSyspqYmV3ufa0CWlGhC4JwnHmI6e8bXmu/SZBkxP\nwwU/2aP8w4viG3s4S+UpE7Zkdre78mS/SUWwb2ShlJFjG7CcBUWrs/E=\n-----END RSA PRIVATE KEY-----\n")
	//SetSignKey(pKey)
	a.config.SignKey = pKey

	pubKey, _ := ReadPublicKey("-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAyqUnBHSpEaaMEqZ2NvTZtK1bC7IPCQKWN8oi0fWZIxqzz4zESTYy\nLdLV/q1cHYVb7hx53QoUNvUUm1SNAqMs4vWI+Dj+hbcnMUusKPB9PQK+WP06STjQ\nCqW6rGzpzftK0tW7leOy7tLSFVMcEvzkKM+04zRYY8+oFh/8rzdHypTIdYWpSyW9\nzHpnhA16yP7TNqZFlWosGLRSG6mfQtK0Mb8KBqsjHah6jzAVltRkfFnPb62xUcQf\nVeaf17wXJN6XFtH5GSEi0snrdiMvVCtnco6NBFzjNeVnh4BJa5+TLvdgE70z2QGu\n8nAV7BJrXBFNFpYTeKumqJ/owgzM2gGo2QIDAQAB\n-----END RSA PUBLIC KEY-----\n")
	//SetVerifyKey(pubKey)
	a.config.VerifyKey = pubKey

	// generate valid token to be used in requests
	claims := customClaims{
		Roles: nil,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   "admin",
			Id:        uuid.New().String(),
			Issuer:    "ORANMGR",
			NotBefore: time.Now().Unix(),
		},
	}
	signedToken, err := generateToken(claims, a.config.Secret, a.config.SignKey)
	if err != nil {
		t.Fatalf("error generating token: %s", err)
	}

	// setup a valid user to be returned by DB mock
	rL, _ := models.NewRoleList([]string{models.AdminRole.String()})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	adminUser := &models.User{
		Username: "admin",
		Version:  0,
		Roles:    rL,
	}
	adminUser.Password = string(hashedPassword)

	// setup tests
	tt := []struct {
		name        string
		status      int
		invalidUser bool
		renReq      *RenewTokenHandlerRequest
	}{
		{
			name:   "invalid renew token and invalid user",
			status: 401,
			renReq: &RenewTokenHandlerRequest{
				UserID:     "notexistinguser",
				RenewToken: "invalidtoken",
			},
			invalidUser: true,
		},
		{
			name:   "invalid renew token and valid user",
			status: 401,
			renReq: &RenewTokenHandlerRequest{
				UserID:     "admin",
				RenewToken: "invalidtoken",
			},
			invalidUser: false,
		},
		{
			name:   "valid renew token and invalid user",
			status: 401,
			renReq: &RenewTokenHandlerRequest{
				UserID:     "notexistinguser",
				RenewToken: signedToken,
			},
			invalidUser: true,
		},
		{
			name:   "valid renew token and valid user",
			status: 200,
			renReq: &RenewTokenHandlerRequest{
				UserID:     "admin",
				RenewToken: signedToken,
			},
			invalidUser: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := bytes.NewBuffer(nil)

			if err := jsonapi.MarshalPayload(requestBody, tc.renReq); err != nil {
				t.Fatalf("could not marshal request body %v", err)
			}

			req, err := http.NewRequest("POST", "/renew", requestBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rec := httptest.NewRecorder()

			s.mock.ExpectBegin()
			s.EventRepo.DB.Begin()

			eQ := s.mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
				WithArgs(sqlmock.AnyArg())
			if tc.invalidUser {
				eQ.WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}))
			} else {
				eQ.WillReturnRows(sqlmock.NewRows([]string{"Username", "Password", "Version"}).
					AddRow(adminUser.Username, adminUser.Password, adminUser.Version))
			}

			a.RenewTokenHandler(rec, req)

			if err := s.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.status {
				t.Errorf("expected status %d; got %d", tc.status, res.StatusCode)
			}
		})
	}
}

func TestSignedTokenHandler(t *testing.T) {
	pKeyString := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAKCAQEAyqUnBHSpEaaMEqZ2NvTZtK1bC7IPCQKWN8oi0fWZIxqzz4zE\nSTYyLdLV/q1cHYVb7hx53QoUNvUUm1SNAqMs4vWI+Dj+hbcnMUusKPB9PQK+WP06\nSTjQCqW6rGzpzftK0tW7leOy7tLSFVMcEvzkKM+04zRYY8+oFh/8rzdHypTIdYWp\nSyW9zHpnhA16yP7TNqZFlWosGLRSG6mfQtK0Mb8KBqsjHah6jzAVltRkfFnPb62x\nUcQfVeaf17wXJN6XFtH5GSEi0snrdiMvVCtnco6NBFzjNeVnh4BJa5+TLvdgE70z\n2QGu8nAV7BJrXBFNFpYTeKumqJ/owgzM2gGo2QIDAQABAoIBAQCmZl0WrJEUPFVz\nDxutXvvSADPt86Wi+WvOnf5fuDOqfse+G1Im6AjmVeWA/mvQlex6JwnudtNImZD1\nR8WOr90w9PwnD+34cQAO25uf9nJwges5+Z49+BflVldmNPz8Nmgnnngtyc7pi1YV\nSqyX7u+Pj5dypk4aj67vlA6S9mrOLk4G8IsJwZt5Hdsp/Erze6wHM7snuZL+acM4\n2FQ7syZ4epvI782y2sRuaiWZMY96CH7PElqWUiz79/asSa1TZaPQ7sDD3AtVqMJm\n5jGeeKwdJJNe+KSLk9gVFnBvfUm38S9UAp/NgjZPFFurGN3ElhZHayK4Q4UL3P/a\n0X2jd7MBAoGBAOn9zgeDNZ3YY0RdAJhyxPGjdER41hb6Gs4Tum/dzRokfygVpCba\nihp42bZqGF1KPXIYZBEky1wYrjOskiiFkICrzJTSMtbHak3+u8hXyYFG0TOZz79y\nvwB4fJp+2G17C7S6uD8a91DZZpQ8dpNWDpvbh1WbZnx3CjbcbJ870YFhAoGBAN20\nkkZpQW5XXQpoPRUuXSljSJk8Gsi8X2yuOTWDFGyDr2VUDAeTd7Jm9PirINOZHD3e\n1rAQ01GMzRdYPC1czULerlRwH3lxN2r3ovU7f1Zu3iCpgrqVSRoeX1R0tisdf62z\nl/V6o1EUxsMEV+Uboccew/hJ0xvNRKLmHmil8MJ5AoGADLp2s6fqibyUocphVumf\nVvmqQHNGSheuz5j5Ik6xcoObuyV6OXbX3lrGlQquapy4PPWgs+IJgegByePQS44A\nb09pIItSoqZUXQvHUT2dQ4ADr0flqidmxnLHbGwL/+CaoWkqzpv76hT5ZITpelhL\nESVe9kQuzgR3tMZGzl6lpeECgYEAt7f+zuJCGlHDA/DFTVwST026x2CLQXT4DnOB\nbNqmfhXRrsIrBcwqEGhI8Be/KBlk0dBrT5Nhyd5HxeSUWXLhlVw6UjZnnpc3OSjk\nnRsktldBMwfFESDMZxxsGuxsWOYk+6grcHykAXiaDNj4jR6MvRi9hG6Ixi0fh23y\nHP4FuOECgYEA3+DT0VORSxntxpYviM6fulJFDhNG8QnqsVlEVCBhjQ2F02yxwMN9\nlHTCzKc9BZMlVhRei0ZqluCVSyspqYmV3ufa0CWlGhC4JwnHmI6e8bXmu/SZBkxP\nwwU/2aP8w4viG3s4S+UpE7Zkdre78mS/SUWwb2ShlJFjG7CcBUWrs/E=\n-----END RSA PRIVATE KEY-----\n"

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pKeyString))

	assert.Nil(t, err)

	/*
		claims := make(jwt.MapClaims)

		claims["dat"] = "hello world"
		claims["exp"] = now.Add(time.Duration(60000000000)).Unix()
		claims["iat"] = now.Unix()
		claims["iss"] = "NONSONOIO"
		claims["nbf"] = now.Unix()
	*/

	var roles []string
	roles = append(roles, "admin")

	claims := customClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(600000)).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   "Pippo",
			Id:        uuid.New().String(),
			Issuer:    "shenok@DESKTOP-1VU82V1",
			NotBefore: time.Now().Unix(),
		},
		Roles: roles,
	}

	_, err = jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)

	assert.Nil(t, err)
}
