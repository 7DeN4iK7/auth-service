package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresCfg PostgresConfig
	ServerCfg   ServerConfig
	JwtCfg      JWTConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type ServerConfig struct {
	Addr              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
}

type JWTConfig struct {
	Secret string
}

func New() (Config, error) {
	config := Config{}

	if err := godotenv.Load(); err != nil {
		return config, err
	}

	config.ServerCfg.Addr = os.Getenv("SERVER_HOST") + ":" + os.Getenv("SERVER_PORT")
	if err := timeEnv(&config.ServerCfg.ReadTimeout, "SERVER_READ_TIMEOUT"); err != nil {
		return config, err
	}

	if err := timeEnv(&config.ServerCfg.ReadHeaderTimeout, "SERVER_READ_HEADER_TIMEOUT"); err != nil {
		return config, err
	}

	if err := timeEnv(&config.ServerCfg.WriteTimeout, "SERVER_WRITE_TIMEOUT"); err != nil {
		return config, err
	}

	if err := timeEnv(&config.ServerCfg.IdleTimeout, "SERVER_IDLE_TIMEOUT"); err != nil {
		return config, err
	}

	config.PostgresCfg.Host = os.Getenv("POSTGRES_HOST")
	config.PostgresCfg.Port = os.Getenv("POSTGRES_PORT")
	config.PostgresCfg.User = os.Getenv("POSTGRES_USER")
	config.PostgresCfg.Password = os.Getenv("POSTGRES_PASSWORD")
	config.PostgresCfg.Database = os.Getenv("POSTGRES_DB")
	config.PostgresCfg.SSLMode = os.Getenv("POSTGRES_SSLMODE")

	config.JwtCfg.Secret = os.Getenv("JWT_SECRET")

	return config, nil
}

func timeEnv(dst *time.Duration, envKey string) error {
	dur, err := time.ParseDuration(os.Getenv(envKey))
	if err != nil {
		return err
	}

	*dst = dur
	return nil
}
