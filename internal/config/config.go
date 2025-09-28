package config

import "os"

type Config struct {
	DevelopmentEnvironment bool
}

func GetConfig() Config {
	appConfig := Config{
		DevelopmentEnvironment: false,
	}
	if devEnv, exists := os.LookupEnv("DEVENV"); exists && devEnv == "1" {
		appConfig.DevelopmentEnvironment = true
	}
	return appConfig
}
