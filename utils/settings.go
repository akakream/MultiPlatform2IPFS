package utils

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	err               error
	baseURL           string
	environment       string
	distromashURL     string
	exportPath        string
	ErrEnvVarNotFound = errors.New("environment variable is not found in the .env file")
)

func InitSettings() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	baseURL, err = GetEnv("BASE_URL", "localhost:3000")
	if err != nil {
		return err
	}
	environment, err = GetEnv("ENVIRONMENT", "DEV")
	if err != nil {
		return err
	}
	exportPath, err = GetEnv("EXPORT_PATH", "./export")
	if err != nil {
		return err
	}
	return nil
}

func GetEnv(envVar string, defaultValue string) (string, error) {
	value, exists := os.LookupEnv(envVar)
	if !exists {
		if defaultValue != "" {
			return "", ErrEnvVarNotFound
		} else {
			return defaultValue, nil
		}
	} else {
		return value, nil
	}
}

func IsEnvDev() bool {
	if environment == "DEV" {
		return true
	} else {
		return false
	}
}
