package main

import (
	"log"
	"net/http"
	"ws/internal/handlers"
)

func main() {
	routes := routes()

	log.Println("starting channel listener")
	go handlers.ListenToWsChannel()
	log.Println("starting server on port 8080")
	if err := http.ListenAndServe(":8080", routes); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
