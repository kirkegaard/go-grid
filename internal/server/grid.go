package server

import (
	"fmt"
	"strconv"
	"sync"
)

// Mutex to protect the grid
var mu sync.Mutex

// Redis key for the grid
const gridKey = "grid_bits"

// 25x25 grid
var gridSize, err = strconv.Atoi(GetEnv("GRID_SIZE", "625"))

func InitGrid() error {
	gridSizeBytes := (gridSize + 7) / 8

	exists, err := rdb.Exists(ctx, gridKey).Result()
	if err != nil {
		return err
	}

	// Return early if the grid already exists
	if exists == 1 {
		return nil
	}

	// Create a zero-filled byte array
	emptyGrid := make([]byte, gridSizeBytes)

	// Set the empty grid in Redis
	err = rdb.Set(ctx, gridKey, emptyGrid, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get the grid state
func getGridState() ([]byte, error) {
	gridSizeBytes := (gridSize + 7) / 8

	// Get the grid state
	gridData, err := rdb.Get(ctx, gridKey).Bytes()

	if err != nil {
		return nil, err
	}

	// If the grid data is the correct size
	if len(gridData) > gridSizeBytes {
		// Truncate the grid data to the expected size
		gridData = gridData[:gridSizeBytes]
	} else if len(gridData) < gridSizeBytes {
		// Pad zeros to the end of the grid data
		gridData = append(gridData, make([]byte, gridSizeBytes-len(gridData))...)
	}

	return gridData, nil
}

// Toggle grid cell
func toggleGridCell(cell int) (int, error) {
	// Lock the grid
	mu.Lock()
	defer mu.Unlock()

	if cell < 0 || cell >= gridSize {
		return -1, fmt.Errorf("cell index out of bounds")
	}

	// Get the current bit value
	currentBit, err := rdb.GetBit(ctx, gridKey, int64(cell)).Result()
	if err != nil {
		return -1, err
	}

	// Toggle the bit
	newBit := 1
	if currentBit == 1 {
		newBit = 0
	}

	// Set the new bit value in redis
	_, err = rdb.SetBit(ctx, gridKey, int64(cell), newBit).Result()
	if err != nil {
		return -1, err
	}

	return newBit, nil
}
