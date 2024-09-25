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
	server.InitGrid()

	mux := http.NewServeMux()

	// Endpoint handlers
	mux.HandleFunc("GET /api/ping", server.PingHandler)
	mux.HandleFunc("GET /api/get", server.ApiGetHandler)
	mux.HandleFunc("POST /api/set", server.ApiSetHandler)
	mux.HandleFunc("GET /api/count", server.ApiGetConnectionsCountHandler)
	mux.HandleFunc("GET /api/clients", server.ApiGetClientsHandler)
	mux.HandleFunc("POST /api/kick", server.ApiKickClientHandler)

	mux.HandleFunc("/api/ws", server.WebSocketHandler)

	// Serve files from the static directory ./www
	mux.HandleFunc("/", server.ApiFallbackHandler)

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
