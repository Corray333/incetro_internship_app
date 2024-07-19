package main

import (
	"fmt"
	"os"

	"github.com/Corray333/internship_app/internal/app"
	"github.com/Corray333/internship_app/internal/config"
	"github.com/Corray333/internship_app/internal/notion"
	"github.com/Corray333/internship_app/internal/storage"
	"github.com/spf13/viper"
)

func main() {
	config.MustInit(os.Args[1])

	fmt.Println("Conf: " + viper.GetString("node_url"))

	fmt.Println()
	fmt.Println("We're getting started...")
	fmt.Println()

	store := storage.New()
	fmt.Println(notion.Sync(store))
	app.New().Run()

}
