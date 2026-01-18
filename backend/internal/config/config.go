// Package config handles application configuration management.
package config

import (
	"os"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	Env        string
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Qdrant     QdrantConfig
	ClickHouse ClickHouseConfig
	NATS       NATSConfig
	Temporal   TemporalConfig
	Blockchain BlockchainConfig
	JWT        JWTConfig
	ML         MLConfig
	Telemetry  TelemetryConfig
}

// ServerConfig holds HTTP/gRPC server settings.
type ServerConfig struct {
	HTTPPort     int
	GRPCPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// QdrantConfig holds Qdrant vector DB settings.
type QdrantConfig struct {
	Host       string
	Port       int
	GRPCPort   int
	Collection string
}

// ClickHouseConfig holds ClickHouse analytics settings.
type ClickHouseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// NATSConfig holds NATS messaging settings.
type NATSConfig struct {
	URL       string
	ClusterID string
	ClientID  string
}

// TemporalConfig holds Temporal workflow engine settings.
type TemporalConfig struct {
	Host      string
	Port      int
	Namespace string
	TaskQueue string
}

// BlockchainConfig holds multi-chain RPC settings.
type BlockchainConfig struct {
	Chains map[string]ChainConfig
}

// ChainConfig holds settings for a single blockchain.
type ChainConfig struct {
	ChainID     int64
	RPCURL      string
	WSURL       string
	BlockTime   time.Duration
	Confirmations int
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret     string
	Expiry     time.Duration
	RefreshExp time.Duration
}

// MLConfig holds ML inference settings.
type MLConfig struct {
	ModelPath      string
	ONNXRuntime    string
	InferenceURL   string
	BatchSize      int
	TimeoutSeconds int
}

// TelemetryConfig holds observability settings.
type TelemetryConfig struct {
	OTELEndpoint   string
	PrometheusPort int
	ServiceName    string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Env: getEnv("VIGILUM_ENV", "development"),
		Server: ServerConfig{
			HTTPPort:     8000,
			GRPCPort:     9000,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: DatabaseConfig{
			Host:         getEnv("POSTGRES_HOST", "localhost"),
			Port:         5432,
			User:         getEnv("POSTGRES_USER", "vigilum"),
			Password:     getEnv("POSTGRES_PASSWORD", ""),
			Database:     getEnv("POSTGRES_DB", "vigilum"),
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: 6379,
			DB:   0,
		},
		Qdrant: QdrantConfig{
			Host:       getEnv("QDRANT_HOST", "localhost"),
			Port:       6333,
			GRPCPort:   6334,
			Collection: "contracts",
		},
		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:     9000,
			User:     "default",
			Database: "vigilum",
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
		Temporal: TemporalConfig{
			Host:      getEnv("TEMPORAL_HOST", "localhost"),
			Port:      7233,
			Namespace: "vigilum",
			TaskQueue: "vigilum-scanner",
		},
		Blockchain: BlockchainConfig{
			Chains: make(map[string]ChainConfig),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiry:     24 * time.Hour,
			RefreshExp: 7 * 24 * time.Hour,
		},
		ML: MLConfig{
			ModelPath:      getEnv("ML_MODEL_PATH", "./ml/models/"),
			ONNXRuntime:    getEnv("ONNX_RUNTIME_PATH", "./ml/onnx/"),
			BatchSize:      32,
			TimeoutSeconds: 30,
		},
		Telemetry: TelemetryConfig{
			OTELEndpoint:   getEnv("OTEL_ENDPOINT", "http://localhost:4317"),
			PrometheusPort: 9090,
			ServiceName:    "vigilum",
		},
	}

	// Load chain configs
	if ethRPC := getEnv("ETH_RPC_URL", ""); ethRPC != "" {
		cfg.Blockchain.Chains["ethereum"] = ChainConfig{
			ChainID:       1,
			RPCURL:        ethRPC,
			BlockTime:     12 * time.Second,
			Confirmations: 12,
		}
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
