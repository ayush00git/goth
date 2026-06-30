package helpers

import (
	"log"
	"os"
)

func GetEnvVar(key string) string {
	value := os.Getenv(key)
	
	if value == "" {
		log.Printf("Error: %s variable missing in environment variables", key)
	}
	return value
}
