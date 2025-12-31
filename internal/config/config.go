package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	MinIO     MinIOConfig
	Qdrant    QdrantConfig
	OpenAI    OpenAIConfig
	Vector    VectorServiceConfig
	Embedding EmbeddingServiceConfig
	Jaeger    JaegerConfig
	Log       LogConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// QdrantConfig Qdrant配置
type QdrantConfig struct {
	Host string
	Port int
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// VectorServiceConfig 向量服务配置
type VectorServiceConfig struct {
	URL string
}

// EmbeddingServiceConfig Embedding服务配置
type EmbeddingServiceConfig struct {
	URL string
}

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	Endpoint string
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string
	Format string
}

// Load 加载配置
func Load() (*Config, error) {
	// 尝试加载.env文件，但不强制要求
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "yoga"),
			Password:        getEnv("DB_PASSWORD", "yoga123"),
			DBName:          getEnv("DB_NAME", "yoga_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY_ID", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_ACCESS_KEY", "minioadmin123"),
			UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET_NAME", "yoga-knowledge"),
		},
		Qdrant: QdrantConfig{
			Host: getEnv("QDRANT_HOST", "localhost"),
			Port: getEnvAsInt("QDRANT_PORT", 6333),
		},
		OpenAI: OpenAIConfig{
			APIKey:  getEnv("OPENAI_API_KEY", ""),
			BaseURL: getEnv("OPENAI_BASE_URL", "https://api.deepseek.com/v1"),
			Model:   getEnv("OPENAI_MODEL", "deepseek-chat"),
		},
		Vector: VectorServiceConfig{
			URL: getEnv("VECTOR_SERVICE_URL", "http://localhost:8003"),
		},
		Embedding: EmbeddingServiceConfig{
			URL: getEnv("EMBEDDING_SERVICE_URL", "http://localhost:8002"),
		},
		Jaeger: JaegerConfig{
			Endpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// validate 验证配置
func (c *Config) validate() error {
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY 未设置（需要 DeepSeek API Key）")
	}
	return nil
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为bool
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为Duration
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// NewLogger 创建日志记录器
func NewLogger(cfg LogConfig) (*zap.Logger, error) {
	var config zap.Config
	if cfg.Level == "debug" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	if cfg.Format == "console" {
		config.Encoding = "console"
	} else {
		config.Encoding = "json"
	}

	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("解析日志级别失败: %w", err)
	}
	config.Level = level

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("构建日志记录器失败: %w", err)
	}

	return logger, nil
}
