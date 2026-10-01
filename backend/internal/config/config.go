package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	AES      AESConfig
	MockAPI  MockAPIConfig
	Worker   WorkerConfig
	QR       QRConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Env  string `mapstructure:"APP_ENV"`
	Port int    `mapstructure:"APP_PORT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     int    `mapstructure:"POSTGRES_PORT"`
	DB       string `mapstructure:"POSTGRES_DB"`
	User     string `mapstructure:"POSTGRES_USER"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	SSLMode  string `mapstructure:"POSTGRES_SSLMODE"`
	MaxConns int32  `mapstructure:"POSTGRES_MAX_CONNS"`
	MinConns int32  `mapstructure:"POSTGRES_MIN_CONNS"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		d.Host, d.Port, d.DB, d.User, d.Password, d.SSLMode, d.MaxConns, d.MinConns,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret          string `mapstructure:"JWT_SECRET"`
	AccessTTLMins   int    `mapstructure:"JWT_ACCESS_TTL_MINUTES"`
	RefreshTTLDays  int    `mapstructure:"JWT_REFRESH_TTL_DAYS"`
}

type AESConfig struct {
	KeyHex string `mapstructure:"AES_KEY_HEX"`
}

type MockAPIConfig struct {
	URL    string `mapstructure:"MOCK_CAMPUS_API_URL"`
	APIKey string `mapstructure:"MOCK_CAMPUS_API_KEY"`
}

type WorkerConfig struct {
	Cron string `mapstructure:"SYNC_WORKER_CRON"`
	Port int    `mapstructure:"SYNC_WORKER_PORT"`
}

type QRConfig struct {
	TOTPPeriodSeconds int `mapstructure:"QR_TOTP_PERIOD_SECONDS"`
	TOTPDigits        int `mapstructure:"QR_TOTP_DIGITS"`
}

type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_SSLMODE", "disable")
	v.SetDefault("POSTGRES_MAX_CONNS", 20)
	v.SetDefault("POSTGRES_MIN_CONNS", 5)
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 60)
	v.SetDefault("JWT_REFRESH_TTL_DAYS", 30)
	v.SetDefault("QR_TOTP_PERIOD_SECONDS", 15)
	v.SetDefault("QR_TOTP_DIGITS", 6)
	v.SetDefault("SYNC_WORKER_CRON", "0 */6 * * *")
	v.SetDefault("SYNC_WORKER_PORT", 8081)

	// Read .env file (optional, environment variables take priority)
	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		v.SetConfigFile("../.env")
		_ = v.ReadInConfig()
	}

	v.AutomaticEnv()

	cfg := &Config{}

	// Map individual keys
	cfg.App.Env = v.GetString("APP_ENV")
	cfg.App.Port = v.GetInt("APP_PORT")

	cfg.Database.Host = v.GetString("POSTGRES_HOST")
	cfg.Database.Port = v.GetInt("POSTGRES_PORT")
	cfg.Database.DB = v.GetString("POSTGRES_DB")
	cfg.Database.User = v.GetString("POSTGRES_USER")
	cfg.Database.Password = v.GetString("POSTGRES_PASSWORD")
	cfg.Database.SSLMode = v.GetString("POSTGRES_SSLMODE")
	cfg.Database.MaxConns = int32(v.GetInt("POSTGRES_MAX_CONNS"))
	cfg.Database.MinConns = int32(v.GetInt("POSTGRES_MIN_CONNS"))

	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = v.GetInt("REDIS_DB")

	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.AccessTTLMins = v.GetInt("JWT_ACCESS_TTL_MINUTES")
	cfg.JWT.RefreshTTLDays = v.GetInt("JWT_REFRESH_TTL_DAYS")

	cfg.AES.KeyHex = v.GetString("AES_KEY_HEX")

	cfg.MockAPI.URL = v.GetString("MOCK_CAMPUS_API_URL")
	cfg.MockAPI.APIKey = v.GetString("MOCK_CAMPUS_API_KEY")

	cfg.Worker.Cron = v.GetString("SYNC_WORKER_CRON")
	cfg.Worker.Port = v.GetInt("SYNC_WORKER_PORT")

	cfg.QR.TOTPPeriodSeconds = v.GetInt("QR_TOTP_PERIOD_SECONDS")
	cfg.QR.TOTPDigits = v.GetInt("QR_TOTP_DIGITS")

	// CORS: split comma-separated string
	originsRaw := v.GetString("CORS_ALLOWED_ORIGINS")
	if originsRaw != "" {
		cfg.CORS.AllowedOrigins = strings.Split(originsRaw, ",")
	}

	return cfg, nil
}
