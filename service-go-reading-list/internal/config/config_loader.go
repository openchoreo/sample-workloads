// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

const (
	DefaultPort     = 8080
	DefaultHostname = "localhost"
)

var (
	EnvName         = "ENV"
	Hostname        = "HOSTNAME"
	Port            = "PORT"
	initialDataPath = "INIT_DATA_PATH"
)

var config Config

func GetConfig() *Config {
	return &config
}

func LoadConfig() (*Config, error) {
	config = Config{
		Hostname:        getEnvString(Hostname, DefaultHostname),
		Port:            getEnvInt(Port, DefaultPort),
		Env:             os.Getenv(EnvName),
		InitialDataPath: os.Getenv(initialDataPath),
	}
	return &config, nil
}

func LoadInitialData() (data InitialData) {
	if config.InitialDataPath == "" {
		return
	}
	contents, err := os.ReadFile(config.InitialDataPath)
	if err != nil {
		log.Fatalf("failed to read initial data at [%s]: %s", config.InitialDataPath, err)
	}
	if err := json.Unmarshal(contents, &data); err != nil {
		log.Fatalf("failed to unmarshal initial data at [%s]: %s", config.InitialDataPath, err)
	}
	return
}

func getEnvInt(key string, defaultVal int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Panic(err)
	}
	return v
}

func getEnvString(key string, defaultVal string) string {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	return s
}
