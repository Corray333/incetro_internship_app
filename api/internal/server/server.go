package server

import (
	"fmt"
	"log/slog"
	"net/http"

	_ "github.com/Corray333/internship_app/docs"
	"github.com/Corray333/internship_app/internal/server/handlers"
	"github.com/Corray333/internship_app/internal/storage"
	"github.com/Corray333/internship_app/internal/telegram"
	"github.com/Corray333/internship_app/pkg/server/auth"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Run(tg *telegram.TelegramClient, store *storage.Storage) {
	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)

	// TODO: get allowed origins, headers and methods from cfg
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Set-Cookie", "Refresh", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Максимальное время кеширования предзапроса (в секундах)
	}))

	router.Group(func(r chi.Router) {
		r.Use(auth.NewAuthMiddleware())

		r.Get("/api/tasks", handlers.ListTasks(store))
		r.Get("/api/tasks/{task_id}", handlers.GetTask(store))
		r.Post("/api/tasks/{task_id}/homework", handlers.SaveHomework(tg, store))
		r.Patch("/api/tasks/{task_id}/homework", handlers.SaveHomework(tg, store))
		r.Patch("/api/tasks/{task_id}", handlers.TaskDone(store))
	})

	router.Get("/api/swagger/*", httpSwagger.WrapHandler)
	router.Post("/api/users/login", handlers.Login(store))
	router.Patch("/api/users/refresh-tokens", handlers.RefreshTokens(store))

	// TODO: add timeouts
	fmt.Println("Port:", viper.GetString("port"))
	server := http.Server{
		Addr:    "0.0.0.0:" + viper.GetString("port"),
		Handler: router,
	}

	slog.Info(server.ListenAndServe().Error())

}
