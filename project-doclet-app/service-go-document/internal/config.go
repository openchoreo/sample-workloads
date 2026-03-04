package document

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	defaultHTTPAddr  = ":8080"
	defaultNATSURL   = "nats://127.0.0.1:4222"
	defaultDBPort    = "5432"
	defaultDBUser    = "doclet"
	defaultDBName    = "doclet"
	defaultDBSSLMode = "disable"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	NATSURL     string
}

func LoadConfig() Config {
	cfg := Config{
		HTTPAddr:    getenv("DOCLET_DOCUMENT_ADDR", defaultHTTPAddr),
		DatabaseURL: loadDatabaseURL(),
		NATSURL:     getenv("DOCLET_NATS_URL", defaultNATSURL),
	}
	return cfg
}

func loadDatabaseURL() string {
	if dbValue := os.Getenv("DOCLET_DATABASE_URL"); dbValue != "" {
		if strings.Contains(dbValue, "://") || strings.Contains(dbValue, "=") {
			return dbValue
		}
		return buildPostgresDSN(dbValue)
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		return buildPostgresDSN(dbHost)
	}
	return ""
}

func buildPostgresDSN(hostValue string) string {
	port := getenv("DB_PORT", defaultDBPort)
	if parsedHost, parsedPort, err := net.SplitHostPort(hostValue); err == nil {
		hostValue = parsedHost
		if os.Getenv("DB_PORT") == "" {
			port = parsedPort
		}
	}
	user := getenvAny([]string{"DB_USER", "POSTGRES_USER"}, defaultDBUser)
	password := getenvAny([]string{"DB_PASSWORD", "POSTGRES_PASSWORD"}, "")
	name := getenvAny([]string{"DB_NAME", "POSTGRES_DB"}, defaultDBName)
	sslmode := getenv("DB_SSLMODE", defaultDBSSLMode)

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		hostValue, port, user, password, name, sslmode,
	)
}

func getenvAny(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
