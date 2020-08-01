package config

import (
	c "github.com/with-go/config"
)

var Config = c.New()

func Get(moduleName string, key string) string {
	return Config.OnModule(moduleName).Get(key)
}

func GetWithDefault(moduleName string, key string, defaultValue string) string {
	return Config.OnModule(moduleName).GetWithDefault(key, defaultValue)
}

func Set(moduleName string, key string, value string) {
	Config.OnModule(moduleName).Set(key, value)
}