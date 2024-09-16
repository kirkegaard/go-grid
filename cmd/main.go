package main

import (
	"fmt"
	"github.com/kirkegaard/go-grid/internal/server"
	"net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func main() {
	server.InitRedis()

	mux := http.NewServeMux()

	// Endpoint handlers
	mux.HandleFunc("GET /ping", server.PingHandler)
	mux.HandleFunc("GET /get", server.ApiGetHandler)
	mux.HandleFunc("POST /set", server.ApiSetHandler)
	mux.HandleFunc("/ws", server.WebSocketHandler)

	hub := server.GetHub()
	go hub.RunHub()

	handler := corsMiddleware(mux)

	// Start the server
	port := server.GetEnv("GRID_PORT", "6060")
	fmt.Printf("Server running on port %s", port)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
	if err != nil {
		panic(err)
	}

}
