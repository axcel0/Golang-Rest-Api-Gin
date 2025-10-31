package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	App      AppConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Mode         string // "debug" or "release"
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string // "debug", "info", "warn", "error"
	Format     string // "json" or "console"
	OutputPath string
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name               string
	Version            string
	Environment        string // "development", "staging", "production"
	RateLimitPerMinute int    // Requests per minute per IP
	RateLimitBurst     int    // Burst size for rate limiter
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  string
	RefreshTokenDuration string
}

// LoadConfig loads configuration from environment and config file using Viper
func LoadConfig() (*Config, error) {
	// Set config file name and path
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Read config file (optional, will use defaults if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; using defaults and environment variables
	}

	// Override with environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.readtimeout", 10*time.Second)
	viper.SetDefault("server.writetimeout", 10*time.Second)
	viper.SetDefault("server.idletimeout", 60*time.Second)

	// Database defaults
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "goproject.db")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.maxidleconns", 10)
	viper.SetDefault("database.maxopenconns", 100)
	viper.SetDefault("database.connmaxlifetime", 1*time.Hour)

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.outputpath", "stdout")

	// App defaults
	viper.SetDefault("app.name", "Go-Lang-project-01")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.ratelimitperminute", 100) // 100 requests per minute
	viper.SetDefault("app.ratelimitburst", 10)      // Allow burst of 10 requests

	// JWT defaults
	viper.SetDefault("jwt.secretkey", "change-this-secret-key-in-production")
	viper.SetDefault("jwt.accesstokenduration", "24h")
	viper.SetDefault("jwt.refreshtokenduration", "168h") // 7 days
}

// GetDSN returns database connection string for PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
