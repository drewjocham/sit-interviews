package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Port    int    `envconfig:"PORT" default:"8081"` // PORT is default to 8080 in App Engine
	Env     string `envconfig:"ENV" default:"development"`
	Version string `envconfig:"VERSION" default:"development"`

	Limiter struct {
		Enabled bool
		Rps     float64
		Burst   int
	}

	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}

	Cors struct {
		TrustedOrigins []string `envconfig:"CORS" default:"http://localhost:5173,default"`
	}

	MongoConfig struct {
		Url              string `envconfig:"DB_URI" default:"mongodb://root:interviews@localhost:27017"`
		DBName           string `envconfig:"DB_NAME" default:"interviewsapi"`
		CourseCollection string `envconfig:"COURSE_COLLECTION" default:"courses"`
		LoginCollection  string `envconfig:"LOGIN_COLLECTION" default:"users"`
		TokenCollection  string `envconfig:"TOKEN_COLLECTION" default:"tokens"`
	}

	UserConfig struct {
		UserCollection string `envconfig:"USER_COLLECTION" default:"users"`
	}

	TokenConfig struct {
		Expires           int           `envconfig:"TOKEN_EXPIRES" default:"24"`
		MaxRetries        int           `envconfig:"TOKEN_CACHE_INIT_MAX_RETRIES" default:"10"`
		RetryPeriod       time.Duration `envconfig:"TOKEN_CACHE_INIT_RETRY_PERIOD" default:"300ms"`
		ValidityThreshold time.Duration `envconfig:"TOKEN_CACHE_VALIDITY_THRESHOLD" default:"3h30m"`
		RefreshPeriod     time.Duration `envconfig:"TOKEN_CACHE_REFRESH_PERIOD" default:"15m"`
	}

	UsersConfig struct {
		UserCollection string `envconfig:"USER_COLLECTION" default:"users"`
	}
}

func NewConfig() (*Config, error) {
	var c Config

	err := envconfig.Process("", &c)

	if err != nil {
		return nil, err
	}

	return &c, nil
}
