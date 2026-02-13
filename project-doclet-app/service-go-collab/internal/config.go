package collab

import "os"

const (
	defaultHTTPAddr = ":8090"
	defaultNATSURL  = "nats://127.0.0.1:4222"
)

type Config struct {
	HTTPAddr string
	NATSURL  string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr: getenv("DOCLET_COLLAB_ADDR", defaultHTTPAddr),
		NATSURL:  getenv("DOCLET_NATS_URL", defaultNATSURL),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
