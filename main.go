package main

import (
	"embed"
	"log"

	"github.com/keainya/service_temp/object"
	"github.com/keainya/service_temp/router"
)

//go:embed web
var webFS embed.FS

func main() {
	if object.Database == nil {
		log.Fatal("database initialization failed")
	}
	log.Println("Account server starting on :8080")
	router.InitRouter(webFS).Run(":8080")
}
