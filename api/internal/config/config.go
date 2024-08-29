package config

import (
	"os"

	"github.com/Corray333/internship_app/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func MustInit(path string) {
	if err := godotenv.Load(path); err != nil {
		panic(err)
	}

	logger.SetupCustomLogger()

	configPath := os.Getenv("CONFIG_PATH")
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
