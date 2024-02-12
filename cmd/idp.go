package main

import (
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"github.com/goidp/config"
	"github.com/goidp/controllers"
	"github.com/goidp/models"
	"io/ioutil"
	stdlog "log"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

func main() {
	envC := config.NewEnvConfigurations()
	c := newControllersConfiguration(envC)

	db, err := models.DBConnect(buildDBConnectionParameters(envC.DB))
	if err != nil {
		log.Fatalf("error initializing database: %s", err)
	}
	location, err := time.LoadLocation(envC.DB.Timezone)
	if err != nil {
		log.Fatalf("failed to parse provided timezone: %s", err.Error())
	}
	if err := models.SetupAutomaticDeletion(db, envC.DB.CleanupPeriod, location, envC.DB.MaxEventsNumber); err != nil {
		log.Fatalf("failed to set-up automatic deletion cronjob for db events: %s", err.Error())
	}
	a := controllers.NewApp(db, c)
	a.AddDefaultUserAndRoles()
	a.Run()
}

func buildDBConnectionParameters(dbC *config.DBConfig) (string, logger.Interface) {
	var password string
	if dbC.Pass == "" {
		data, err := ioutil.ReadFile(dbC.PassFile)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalf("unable to read DB password file")
		}
		password = strings.Trim(string(data), "\n\t ")
	} else {
		password = dbC.Pass
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", dbC.Host, dbC.User, password, dbC.Name, dbC.Port, dbC.SSLMode, dbC.Timezone)
	var logLevel logger.LogLevel
	if dbC.ShowSql {
		logLevel = logger.Info
	} else {
		logLevel = logger.Silent
	}

	return dsn, logger.New(
		stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: 5,        // Slow SQL threshold
			LogLevel:      logLevel, // Log level
		},
	)
}

func newControllersConfiguration(c *config.EnvConfigurations) *controllers.Config {
	var verifyKey *rsa.PublicKey
	var signKey *rsa.PrivateKey
	var secret string

	// Utility function to read a file from the FS if it exists, returns the passed
	// string otherwise
	readIfFileFunction := func(in string) []byte {
		if _, err := os.Stat(in); err == nil {
			if content, err := os.ReadFile(in); err != nil {
				log.Fatalf("error when reading file '%s': %v", in, err.Error())
			} else {
				return content
			}
		}

		// It is not a file, it is passed 'by value'
		return []byte(in)
	}

	if c.JWT.UseKey {

		// If keys are configured read them from the filesystem, decode them as they
		// expected to be in PEM format and parse/read them thereafter.
		pubKeyData, _ := pem.Decode(readIfFileFunction(c.JWT.PublicKey))
		pvtKeyData, _ := pem.Decode(readIfFileFunction(c.JWT.PrivateKey))

		if pubKeyData == nil {
			log.Fatalf("public key not in PEM format")
		}
		if pvtKeyData == nil {
			log.Fatalf("private key not in PEM format")
		}
		x509PublicKey := controllers.IsX509Certificate(pubKeyData.Bytes)
		if x509PublicKey {
			log.Infof("Public key format is X509")
		} else {
			log.Infof("Public key format is RSA")
		}
		var err error
		verifyKey, signKey, err = controllers.ReadKeyPair(pubKeyData.Bytes,
			pvtKeyData.Bytes,
			x509PublicKey)
		if err != nil {
			log.Fatalf("error when extracting keys: %s", err)
		}
	} else {
		secret = string(readIfFileFunction(c.JWT.Secret))
	}
	var renewTokenExpireTime time.Duration
	if c.JWT.Refresh {
		renewTokenExpireTime = c.JWT.RefreshExpireTime
	}
	pKeys := controllers.ReadPublicKeys(c.JWT.PublicKeysPath)
	cC := controllers.Config{
		LogLevel:              c.App.LogLevel,
		WriteTimeout:          c.App.WriteTimeout,
		ReadTimeout:           c.App.ReadTimeout,
		IdleTimeout:           c.App.IdleTimeout,
		Host:                  c.App.Host,
		Port:                  c.App.Port,
		Secret:                secret,
		SignKey:               signKey,
		VerifyKey:             verifyKey,
		AccessTokenExpireTime: c.JWT.AccessExpireTime,
		RenewTokenExpireTime:  renewTokenExpireTime,
		TrustedPublicKeys:     pKeys,
	}
	return &cC
}
