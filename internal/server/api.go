package server

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
)

// Fetch the grid state
func ApiGetHandler(w http.ResponseWriter, r *http.Request) {
	bits := getGridState()
	compressed := hex.EncodeToString(bits)

	w.Header().Set("Content-Type", "plain/text")
	fmt.Fprint(w, compressed)
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
