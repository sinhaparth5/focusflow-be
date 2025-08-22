package config

import (
	"os"
)

type Config struct {
	Port                string
	FirebaseAPIKey      string
	FirebaseAuthDomain  string
	FirebaseProjectID   string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURI   string
	JWTSecret          string
}

func New() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		FirebaseAPIKey:      getEnv("FIREBASE_API_KEY", ""),
		FirebaseAuthDomain:  getEnv("FIREBASE_AUTH_DOMAIN", ""),
		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", ""),
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:   getEnv("GOOGLE_REDIRECT_URI", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}