package server

import (
	"sync"
)

// Mutex to protect the grid
var mu sync.Mutex

// Redis key for the grid
const gridKey = "grid_bits"

// 25x25 grid
const gridSize = 626

func getGridState() []byte {
	// Create a byte slice to hold the bits as bytes.
	// Each byte will hold 8 bits of the grid.
	// The 7 in (gridSize+7)/8 is to round up the division result.
	bits := make([]byte, (gridSize+7)/8)

	for i := 0; i < gridSize; i++ {
		bit, err := rdb.GetBit(ctx, gridKey, int64(i)).Result()
		if err != nil {
			continue
		}

		if bit == 1 {
			// Set the bit in the byte slice
			// The bit index in the byte is the remainder of i divided by 8
			bits[i/8] ^= 1 << (i % 8)
		}
	}

	return bits
}

// Toggle grid cell handler
func toggleGridCell(cell int) (int, error) {
	// Lock the grid
	mu.Lock()
	defer mu.Unlock()

	// GetBit returns an int64, so we need to convert it to an int
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
