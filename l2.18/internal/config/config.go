package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"calendar-wbf/internal/entity"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App    App    `env-prefix:"APP_"`
		Logger Logger `env-prefix:"LOGGER_"`
		HTTP   HTTP   `env-prefix:"HTTP_"`
		Cache  Cache  `env-prefix:"CACHE_"`
		Env    string `                     env:"ENV" env-default:"local" validate:"oneof=local dev staging prod"`
	}

	App struct {
		Port    int    `env:"PORT"    validate:"gte=1,lte=65535" env-default:"8080"`
		Name    string `env:"NAME"    validate:"required"`
		Version string `env:"VERSION" validate:"required"`
	}

	HTTP struct {
		Host              string        `env:"HOST"                validate:"required"                 env-default:"0.0.0.0"`
		Port              string        `env:"PORT"                validate:"required,gte=1,lte=65535" env-default:"8080"`
		ReadTimeout       time.Duration `env:"READ_TIMEOUT"        validate:"gte=10ms,lte=30s"         env-default:"5s"`
		WriteTimeout      time.Duration `env:"WRITE_TIMEOUT"       validate:"gte=10ms,lte=30s"         env-default:"5s"`
		IdleTimeout       time.Duration `env:"IDLE_TIMEOUT"        validate:"gte=10ms,lte=30s"         env-default:"60s"`
		ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT"    validate:"gte=10ms,lte=30s"         env-default:"10s"`
		ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" validate:"gte=10ms,lte=30s"         env-default:"5s"`
	}

	Cache struct {
		Capacity        int           `env:"CAPACITY"         validate:"required,min=1,max=1000000"`
		TTL             time.Duration `env:"TTL"              validate:"required,gt=0s,lte=24h"     env-default:"5m"`
		CleanupInterval time.Duration `env:"CLEANUP_INTERVAL" validate:"gt=0s,lte=24h"              env-default:"10s"`
	}

	Logger struct {
		Level      string `env:"LEVEL"       env-default:"info"                     validate:"oneof=debug info warn error"`
		Filename   string `env:"FILENAME"    env-default:"./logs/order-service.log"`
		MaxSize    int    `env:"MAX_SIZE"    env-default:"100"                      validate:"min=1,max=1000"`
		MaxBackups int    `env:"MAX_BACKUPS" env-default:"3"                        validate:"min=0,max=20"`
		MaxAge     int    `env:"MAX_AGE"     env-default:"28"                       validate:"min=1,max=365"`
	}
)

func Load() (*Config, error) {
	path := fetchConfigPath()
	if path == "" {
		return nil, entity.ErrConfigPathNotSet
	}
	return LoadPath(path)
}

func LoadPath(configPath string) (*Config, error) {
	const op = "config.LoadPath"

	validate := validator.New()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: config file does not exist: %s", op, configPath)
	} else if err != nil {
		return nil, fmt.Errorf("%s: checking config file: %w", op, err)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("%s: read config: %w", op, err)
	}

	var validationErrors []string
	if err := validate.Struct(&cfg); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, ve := range validationErrs {
				validationErrors = append(validationErrors,
					fmt.Sprintf("%s=%v must satisfy '%s'", ve.Field(), ve.Value(), ve.Tag()))
			}
			return nil, fmt.Errorf(
				"%s: config validation: %v", op,
				strings.Join(validationErrors, "; "),
			)
		}
		return nil, fmt.Errorf("%s: config validation: %w", op, err)
	}

	return &cfg, nil
}

func fetchConfigPath() string {
	var path string
	flag.StringVar(&path, "config", "", "Path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	return path
}
