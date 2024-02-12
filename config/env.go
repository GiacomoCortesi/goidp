package config

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
)

type EnvConfigurations struct {
	DB  *DBConfig
	JWT *JWTConfig
	App *AppConfig
}

func NewEnvConfigurations() *EnvConfigurations {
	var c EnvConfigurations
	c.LoadEnvConfigurations()
	return &c
}

func (eC *EnvConfigurations) LoadEnvConfigurations() {
	var jC JWTConfig
	var dbC DBConfig
	var appC AppConfig

	err := envconfig.Process("jwt", &jC)
	if err != nil {
		log.Fatal(err.Error())
	}
	eC.JWT = &jC

	err = envconfig.Process("db", &dbC)
	if err != nil {
		log.Fatal(err.Error())
	}
	eC.DB = &dbC

	err = envconfig.Process("app", &appC)
	if err != nil {
		log.Fatal(err.Error())
	}
	eC.App = &appC
}

type AppConfig struct {
	Host         string `default:"0.0.0.0"`
	Port         string `default:"8889"`
	WriteTimeout int    `default:"15" split_words:"true"`
	ReadTimeout  int    `default:"15" split_words:"true"`
	IdleTimeout  int    `default:"60" split_words:"true"`
	LogLevel     string `default:"debug" split_words:"true"`
}

type DBConfig struct {
	Name            string `default:"idp"`
	User            string `default:"postgres"`
	PassFile        string `default:"/run/secrets/db_secret" split_words:"true"`
	Pass            string `default:""`
	SSLMode         string `default:"disable" split_words:"true"`
	Timezone        string `default:"UTC"`
	Host            string `default:"idp-db"`
	Port            string `default:"5432"`
	Charset         string `default:"utf8"`
	ParseTime       bool   `default:"true" split_words:"true"`
	ShowSql         bool   `default:"true" split_words:"true"`
	CleanupPeriod   string `default:"0 2 * * *" split_words:"true"`
	MaxEventsNumber int64  `default:"100" split_words:"true"`
}

type JWTConfig struct {
	PublicKey         string        `default:"/run/secrets/jwt_public" split_words:"true"`
	PrivateKey        string        `default:"/run/secrets/jwt_private" split_words:"true"`
	AccessExpireTime  time.Duration `default:"5m" split_words:"true"`
	RefreshExpireTime time.Duration `default:"2400m" split_words:"true"`
	Refresh           bool          `default:"True"`
	UseKey            bool          `default:"True"`
	Secret            string        `default:""`
	PublicKeysPath    string        `default:"/run/pubkeys" split_words:"true"`
}
