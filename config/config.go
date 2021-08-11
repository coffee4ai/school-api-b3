package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type appConfig struct {
	ApiPort    string
	MongoDBURL string
	ApiSecret  string
	MongoDB    string
	B3BaseURL  string
	B3Salt     string
}

var defaultConfig appConfig

func LoadConfig() error {
	if godotenv.Load() != nil {
		return fmt.Errorf("error loading the .env file")
	}

	if defaultConfig.ApiPort = os.Getenv("API_PORT"); defaultConfig.ApiPort == "" {
		defaultConfig.ApiPort = "9696"
		// return fmt.Errorf("API Port not defined in .env file")
	}

	if defaultConfig.ApiSecret = os.Getenv("API_SECRET"); defaultConfig.ApiSecret == "" {
		defaultConfig.ApiSecret = "$ch0oLAPi4B#CUBE@Web%%"
		// return fmt.Errorf("API Port not defined in .env file")
	}

	//how best to check if any of the DB param is not available?
	defaultConfig.MongoDBURL = fmt.Sprintf("%s+srv://%s:%s@%s/%s?retryWrites=true&w=majority",
		os.Getenv("DB_DRIVER"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"))

	defaultConfig.MongoDB = os.Getenv("DB_NAME")
	defaultConfig.B3BaseURL = os.Getenv("B3BASEURL")
	defaultConfig.B3Salt = os.Getenv("B3SALT")

	return nil
}

func GetApiPort() string {
	return defaultConfig.ApiPort
}

func GetApiSecret() string {
	return defaultConfig.ApiSecret
}

func GetMongoDbUrl() string {
	return defaultConfig.MongoDBURL
}

func GetMongoDbName() string {
	return defaultConfig.MongoDB
}

func GetB3BaseURL() string {
	return defaultConfig.B3BaseURL
}

func GetB3Salt() string {
	return defaultConfig.B3Salt
}
