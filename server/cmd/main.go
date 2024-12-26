package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"server/internal/server"
	"server/internal/server/clients"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
)

func main() {
	flag.Parse()

	hub := server.NewHub()

	// whenever a client hits 'ws://localhost:8080/ws' this function will run
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// NewWebSocketClient function returns a new client
		hub.Serve(clients.NewWebSocketClient, w, r)
	})
	go hub.Run()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)

	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Error starting server %s", err.Error())
	}
}
