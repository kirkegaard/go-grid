package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func ApiFallbackHandler(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir("./www")).ServeHTTP(w, r)
}

func ApiGetConnectionsCountHandler(w http.ResponseWriter, r *http.Request) {
	clients := hub.GetConnectedClients()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(len(clients))
}

func ApiGetClientsHandler(w http.ResponseWriter, r *http.Request) {
	clients := hub.GetConnectedClients()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func ApiKickClientHandler(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("clientId")
	if clientId == "" {
		http.Error(w, "Missing clientId", http.StatusBadRequest)
		return
	}

	duration, err := strconv.Atoi(r.URL.Query().Get("duration"))
	if err != nil || duration == 0 {
		duration = 0
	}

	kicked := hub.KickClient(clientId, duration)
	if !kicked {
		http.Error(w, "Client not found", http.StatusBadRequest)
		return
	}
	http.Error(w, "Client kicked", http.StatusOK)
}

// Fetch the grid state
func ApiGetHandler(w http.ResponseWriter, r *http.Request) {
	bits, err := getGridState()

	if err != nil {
		http.Error(w, "Failed to get grid state", http.StatusInternalServerError)
		return
	}

	based := base64.StdEncoding.EncodeToString(bits)

	w.Header().Set("Content-Type", "plain/text")
	fmt.Fprint(w, based)
}

// Set a cell in the grid
func ApiSetHandler(w http.ResponseWriter, r *http.Request) {
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

	newBit, err := toggleGridCell(cell)
	if err != nil {
		http.Error(w, "Failed to toggle grid cell", http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("%d:%d", cell, newBit)
	hub.Broadcast <- []byte(message)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, message)
}
