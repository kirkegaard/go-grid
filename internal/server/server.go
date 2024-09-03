package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
	"sync"
)

// Mutex to protect the grid
var mu sync.Mutex

// Redis says i need context so i gave it context
var ctx = context.Background()

// Redis client
var rdb *redis.Client

// Redis key for the grid
const gridKey = "grid_bits"

// 25x25 grid
const gridSize = 625

func Run() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	// Endpoint handlers
	mux.HandleFunc("GET /ping", ping)
	mux.HandleFunc("GET /get", getGridHandler)
	mux.HandleFunc("POST /set", setGridHandler)

	handler := corsMiddleware(mux)

	// Start the server
	fmt.Println("Server running on port 6060")
	err = http.ListenAndServe(":6060", handler)
	if err != nil {
		panic(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

// Fetch grid handler
func getGridHandler(w http.ResponseWriter, r *http.Request) {
	// Create a byte slice to hold the bits as bytes
	bits := make([]byte, (gridSize+7)/8)

	for i := 0; i < gridSize; i++ {
		bit, err := rdb.GetBit(ctx, gridKey, int64(i)).Result()
		if err != nil {
			http.Error(w, "Error fetching grid data", http.StatusInternalServerError)
			return
		}

		if bit == 1 {
			bits[i/8] ^= 1 << (i % 8)
		}
	}

	compressed := hex.EncodeToString(bits)

	w.Header().Set("Content-Type", "plain/text")
	fmt.Fprint(w, compressed)
}

// Toggle grid cell handler
func setGridHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data (cell index)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	cellStr := r.FormValue("cell")
	if cellStr == "" {
		http.Error(w, "Cell parameter is missing", http.StatusBadRequest)
		return
	}

	// Convert the cell index to an integer
	cell, err := strconv.Atoi(cellStr)
	// Check if the cell index is valid
	if err != nil || cell < 0 || cell >= gridSize {
		http.Error(w, "Invalid cell index", http.StatusBadRequest)
		return
	}

	// GetBit returns an int64, so we need to convert it to an int
	currentBit, err := rdb.GetBit(ctx, gridKey, int64(cell)).Result()
	if err != nil {
		http.Error(w, "Error fetching grid data", http.StatusInternalServerError)
		return
	}

	// Toggle the bit
	newBit := 1
	if currentBit == 1 {
		newBit = 0
	}

	// Lock the grid
	mu.Lock()
	defer mu.Unlock()

	// Set the new bit value in redis
	_, err = rdb.SetBit(ctx, gridKey, int64(cell), newBit).Result()
	if err != nil {
		http.Error(w, "Error updating grid data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Bit at cell %d toggled successfully", cell)
}
