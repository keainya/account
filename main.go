package main

import (
	"embed"
	"log"

	"github.com/keainya/account/object"
	"github.com/keainya/account/router"
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
