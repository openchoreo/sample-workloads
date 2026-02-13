package document

import "os"

const (
	defaultHTTPAddr = ":8080"
	defaultNATSURL  = "nats://127.0.0.1:4222"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	NATSURL     string
}

func LoadConfig() Config {
	cfg := Config{
		HTTPAddr:    getenv("DOCLET_DOCUMENT_ADDR", defaultHTTPAddr),
		DatabaseURL: os.Getenv("DOCLET_DATABASE_URL"),
		NATSURL:     getenv("DOCLET_NATS_URL", defaultNATSURL),
	}
	return cfg
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
