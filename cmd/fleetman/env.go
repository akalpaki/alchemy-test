package main

import (
	"os"
	"strconv"

	"github.com/akalpaki/alchemy-test/internal/config"
	"github.com/joho/godotenv"
)

func loadConfig() *config.Config {
	godotenv.Load(".env")
	logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if err != nil {
		panic("log level provided is not a number!")
	}
	return config.New(
		config.WithConnStr(os.Getenv("CONNECTION_STRING")),
		config.WithLogLevel(logLevel),
		config.WithLogFile(os.Getenv("LOG_FILE")),
	)
}
