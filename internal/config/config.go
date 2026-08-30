package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	LiAt       string
	JSessionID string
	Email      string
	Password   string

	APIKey          string
	CacheTTLSeconds int
	Port            string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		LiAt:            os.Getenv("LINKEDIN_LI_AT"),
		JSessionID:      os.Getenv("LINKEDIN_JSESSIONID"),
		Email:           os.Getenv("LINKEDIN_EMAIL"),
		Password:        os.Getenv("LINKEDIN_PASSWORD"),
		APIKey:          os.Getenv("API_KEY"),
		CacheTTLSeconds: envInt("CACHE_TTL_SECONDS", 3600),
		Port:            envStr("PORT", "8000"),
	}
}

func (c *Config) HasCookieAuth() bool { return c.LiAt != "" }

func (c *Config) HasCredentialAuth() bool { return c.Email != "" && c.Password != "" }

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
