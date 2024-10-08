package main

import (
	"os"

	"github.com/Corray333/internship_app/internal/app"
	"github.com/Corray333/internship_app/internal/config"
	"github.com/Corray333/internship_app/internal/notion"
	"github.com/Corray333/internship_app/internal/storage"
)

func main() {
	config.MustInit(os.Args[1])

	// fmt.Println("Conf: " + viper.GetString("node_url"))

	// fmt.Println()
	// fmt.Println("We're getting started...")
	// fmt.Println()

	store := storage.New()
	go notion.Sync(store)
	app.New().Run()

}
