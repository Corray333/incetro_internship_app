package app

import (
	"os"

	"github.com/Corray333/internship_app/internal/server"
	"github.com/Corray333/internship_app/internal/storage"
	"github.com/Corray333/internship_app/internal/telegram"
)

type App struct {
	tg *telegram.TelegramClient
	// storage Storage
}

func New() *App {
	store := storage.New()
	return &App{
		tg: telegram.NewClient(os.Getenv("BOT_TOKEN"), store),
		// storage: store,
	}
}

func (app *App) Run() {
	store := storage.New()

	go app.tg.Run()
	server.Run(app.tg, store)

}
