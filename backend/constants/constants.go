package constants

import "os"

const (
	USER_INDEX    = "user"
	POST_INDEX    = "post"
	SESSION_INDEX = "session"
)

var (
	ES_URL      = env("ES_URL", "http://127.0.0.1:9200")
	ES_USERNAME = os.Getenv("ES_USERNAME")
	ES_PASSWORD = os.Getenv("ES_PASSWORD")
	GCS_BUCKET  = os.Getenv("GCS_BUCKET")
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
