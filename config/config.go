package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Log       LogConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	Scheduler SchedulerConfig
	Worker    WorkerConfig
	Session   SessionConfig
	Resend    ResendConfig
}

type AppConfig struct{ Name, Env string }
type HTTPConfig struct{ Addr string }
type LogConfig struct {
	Dir, Level, Timezone string
	RetentionDays        int
	Console, AddSource   bool
}
type MySQLConfig struct {
	Enabled          bool
	Driver, DSN      string
	MaxOpen, MaxIdle int
}
type RedisConfig struct {
	Enabled        bool
	Addr, Password string
	DB             int
}
type SchedulerConfig struct {
	Enabled  bool
	Timezone string
}
type WorkerConfig struct {
	Enabled          bool
	Count, QueueSize int
}
type SessionConfig struct {
	CookieName string
	TTLMinutes int
	Secure     bool
}
type ResendConfig struct {
	Enabled      bool
	APIKey, From string
}

func Load() (Config, error) {
	loadDotEnv(".env")
	retention, err := intEnv("LOG_RETENTION_DAYS", 30)
	if err != nil {
		return Config{}, err
	}
	maxOpen, err := intEnv("MYSQL_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	maxIdle, err := intEnv("MYSQL_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := intEnv("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}
	workerCount, err := intEnv("WORKER_COUNT", 2)
	if err != nil {
		return Config{}, err
	}
	queueSize, err := intEnv("WORKER_QUEUE_SIZE", 32)
	if err != nil {
		return Config{}, err
	}
	return Config{
		App:       AppConfig{env("APP_NAME", "qesyweb"), env("APP_ENV", "development")},
		HTTP:      HTTPConfig{env("HTTP_ADDR", ":8080")},
		Log:       LogConfig{env("LOG_DIR", "./logs"), env("LOG_LEVEL", "info"), env("LOG_TIMEZONE", time.Local.String()), retention, boolEnv("LOG_CONSOLE", true), boolEnv("LOG_ADD_SOURCE", false)},
		MySQL:     MySQLConfig{boolEnv("MYSQL_ENABLED", false), env("MYSQL_DRIVER", "mysql"), env("MYSQL_DSN", ""), maxOpen, maxIdle},
		Redis:     RedisConfig{boolEnv("REDIS_ENABLED", false), env("REDIS_ADDR", "127.0.0.1:6379"), env("REDIS_PASSWORD", ""), redisDB},
		Scheduler: SchedulerConfig{boolEnv("SCHEDULER_ENABLED", true), env("SCHEDULER_TIMEZONE", "Asia/Shanghai")},
		Worker:    WorkerConfig{boolEnv("WORKER_ENABLED", true), workerCount, queueSize},
		Session:   SessionConfig{env("SESSION_COOKIE_NAME", "qesyweb_session"), mustInt("SESSION_TTL_MINUTES", 120), boolEnv("SESSION_COOKIE_SECURE", false)},
		Resend:    ResendConfig{boolEnv("RESEND_ENABLED", false), env("RESEND_API_KEY", ""), env("RESEND_FROM", "")},
	}, nil
}

func mustInt(key string, fallback int) int {
	value, err := intEnv(key, fallback)
	if err != nil {
		return fallback
	}
	return value
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		value = strings.Trim(value, "\"'")
		if key != "" {
			_ = os.Setenv(key, value)
		}
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func intEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
func boolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}
