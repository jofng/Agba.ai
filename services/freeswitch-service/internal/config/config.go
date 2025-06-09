package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the FreeSWITCH service
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	NATS       NATSConfig       `mapstructure:"nats"`
	FreeSWITCH FreeSWITCHConfig `mapstructure:"freeswitch"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Services   ServicesConfig   `mapstructure:"services"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL       string `mapstructure:"url"`
	ClusterID string `mapstructure:"cluster_id"`
	ClientID  string `mapstructure:"client_id"`
}

// FreeSWITCHConfig holds FreeSWITCH specific configuration
type FreeSWITCHConfig struct {
	Host              string            `mapstructure:"host"`
	Port              int               `mapstructure:"port"`
	Password          string            `mapstructure:"password"`
	EventSocketHost   string            `mapstructure:"event_socket_host"`
	EventSocketPort   int               `mapstructure:"event_socket_port"`
	ConfigPath        string            `mapstructure:"config_path"`
	LogPath           string            `mapstructure:"log_path"`
	RecordingsPath    string            `mapstructure:"recordings_path"`
	SIPProfiles       []SIPProfile      `mapstructure:"sip_profiles"`
	DefaultTrunk      string            `mapstructure:"default_trunk"`
	MaxConcurrentCalls int              `mapstructure:"max_concurrent_calls"`
	CallTimeout       int               `mapstructure:"call_timeout"`
	Variables         map[string]string `mapstructure:"variables"`
}

// SIPProfile represents a SIP profile configuration
type SIPProfile struct {
	Name     string            `mapstructure:"name"`
	Host     string            `mapstructure:"host"`
	Port     int               `mapstructure:"port"`
	Settings map[string]string `mapstructure:"settings"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// ServicesConfig holds configuration for other services
type ServicesConfig struct {
	RecordingService    ServiceEndpoint `mapstructure:"recording_service"`
	MonitoringService   ServiceEndpoint `mapstructure:"monitoring_service"`
	UserService         ServiceEndpoint `mapstructure:"user_service"`
	ConfigService       ServiceEndpoint `mapstructure:"config_service"`
	NotificationService ServiceEndpoint `mapstructure:"notification_service"`
}

// ServiceEndpoint represents a service endpoint configuration
type ServiceEndpoint struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	Timeout int    `mapstructure:"timeout"`
	Retries int    `mapstructure:"retries"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Set default values
	setDefaults()

	// Load from config file if exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive data
	overrideWithEnv(config)

	return config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.dbname", "agba_freeswitch")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 300)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	// NATS defaults
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.cluster_id", "agba-cluster")
	viper.SetDefault("nats.client_id", "freeswitch-service")

	// FreeSWITCH defaults
	viper.SetDefault("freeswitch.host", "localhost")
	viper.SetDefault("freeswitch.port", 5060)
	viper.SetDefault("freeswitch.event_socket_host", "localhost")
	viper.SetDefault("freeswitch.event_socket_port", 8021)
	viper.SetDefault("freeswitch.password", "ClueCon")
	viper.SetDefault("freeswitch.config_path", "/usr/local/freeswitch/conf")
	viper.SetDefault("freeswitch.log_path", "/usr/local/freeswitch/log")
	viper.SetDefault("freeswitch.recordings_path", "/usr/local/freeswitch/recordings")
	viper.SetDefault("freeswitch.max_concurrent_calls", 1000)
	viper.SetDefault("freeswitch.call_timeout", 60)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)

	// Service defaults
	viper.SetDefault("services.recording_service.host", "recording-service")
	viper.SetDefault("services.recording_service.port", 8080)
	viper.SetDefault("services.recording_service.timeout", 30)
	viper.SetDefault("services.recording_service.retries", 3)

	viper.SetDefault("services.monitoring_service.host", "monitoring-service")
	viper.SetDefault("services.monitoring_service.port", 8080)
	viper.SetDefault("services.monitoring_service.timeout", 30)
	viper.SetDefault("services.monitoring_service.retries", 3)

	viper.SetDefault("services.user_service.host", "user-management-service")
	viper.SetDefault("services.user_service.port", 8080)
	viper.SetDefault("services.user_service.timeout", 30)
	viper.SetDefault("services.user_service.retries", 3)

	viper.SetDefault("services.config_service.host", "config-service")
	viper.SetDefault("services.config_service.port", 8080)
	viper.SetDefault("services.config_service.timeout", 30)
	viper.SetDefault("services.config_service.retries", 3)

	viper.SetDefault("services.notification_service.host", "notification-service")
	viper.SetDefault("services.notification_service.port", 8080)
	viper.SetDefault("services.notification_service.timeout", 30)
	viper.SetDefault("services.notification_service.retries", 3)
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) {
	// Database password
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}

	// Redis password
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}

	// FreeSWITCH password
	if fsPassword := os.Getenv("FREESWITCH_PASSWORD"); fsPassword != "" {
		config.FreeSWITCH.Password = fsPassword
	}

	// Server port
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetFreeSWITCHEventSocketAddr returns the FreeSWITCH Event Socket address
func (c *Config) GetFreeSWITCHEventSocketAddr() string {
	return fmt.Sprintf("%s:%d", c.FreeSWITCH.EventSocketHost, c.FreeSWITCH.EventSocketPort)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.FreeSWITCH.Host == "" {
		return fmt.Errorf("freeswitch host is required")
	}

	if c.FreeSWITCH.Password == "" {
		return fmt.Errorf("freeswitch password is required")
	}

	return nil
}