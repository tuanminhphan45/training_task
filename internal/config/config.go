package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DB     DBConfig
	Server ServerConfig
	Crawl  CrawlConfig
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type ServerConfig struct {
	Port string
}

type CrawlConfig struct {
	BaseURL       string
	MaxFiles      int
	MaxConcurrent int
	BatchSize     int
	OutDir        string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "virusshare"),
			Password:        getEnv("DB_PASSWORD", "virusshare123"),
			Name:            getEnv("DB_NAME", "virusshare_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Crawl: CrawlConfig{
			BaseURL:       getEnv("CRAWL_BASE_URL", "https://virusshare.com/hashfiles/"),
			MaxFiles:      getEnvInt("CRAWL_MAX_FILES", 499),
			MaxConcurrent: getEnvInt("CRAWL_MAX_CONCURRENT", 50),
			BatchSize:     getEnvInt("CRAWL_BATCH_SIZE", 1000),
			OutDir:        getEnv("CRAWL_OUT_DIR", "data/hashfiles"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
