package main

import (
	"log"

	"github.com/keainya/service_temp/object"
	"github.com/keainya/service_temp/router"
)

func main() {
	if object.Database == nil {
		log.Fatal("database initialization failed")
	}
	log.Println("Account server starting on :8080")
	router.InitRouter().Run(":8080")
}
