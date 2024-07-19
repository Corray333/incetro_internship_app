package config

import (
	"log/slog"
	"os"

	"github.com/Corray333/internship_app/pkg/server/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func MustInit(path string) {
	if err := godotenv.Load(path); err != nil {
		panic(err)
	}

	setupLogger()

	configPath := os.Getenv("CONFIG_PATH")
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func setupLogger() {
	handler := logger.NewHandler(nil)
	log := slog.New(handler)
	slog.SetDefault(log)
}
