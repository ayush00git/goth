package helpers

import (
	"os"
	"fmt"
	"github.com/joho/godotenv"
)

func GetEnvVar(env string) string {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading the env variables %s", err)
		os.Exit(1)
	}

	key := os.Getenv(env)
	if key == "" {
		fmt.Printf("No env variable with key = %s found", key)
		os.Exit(1)
	}
	return key
}
