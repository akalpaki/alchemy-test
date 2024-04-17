package config

import (
	"log/slog"
	"os"
	"slices"
)

// these values are equivalent to those from the slog.Level default levels
var validLogLevels = []int{-4, 0, 4, 8}

type option func(*Config)

type Config struct {
	ConnStr  string
	LogLevel slog.Level
	LogFile  string
}

func New(opts ...option) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithConnStr(connStr string) option {
	return func(c *Config) {
		c.ConnStr = connStr
	}
}

func WithLogLevel(logLevel int) option {
	if !slices.Contains(validLogLevels, logLevel) {
		panic("invalid log level value given")
	}
	lvl := slog.Level(logLevel)
	return func(c *Config) {
		c.LogLevel = lvl
	}
}

func WithLogFile(logFile string) option {
	if logFile == "" {
		logFile = os.Stdout.Name()
	}
	return func(c *Config) {
		c.LogFile = logFile
	}
}
