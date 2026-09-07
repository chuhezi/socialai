package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"socialai/backend"
	"socialai/handler"
)

func main() {
	fmt.Println("started-service")

	backend.InitElasticsearchBackend()
	backend.InitGCSBackend()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: net.JoinHostPort(os.Getenv("HOST"), port), Handler: handler.InitRouter(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 90 * time.Second, IdleTimeout: 90 * time.Second}
	log.Fatal(server.ListenAndServe())
}
