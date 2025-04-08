package models

import (
	"github.com/oklog/ulid/v2"
	"math/rand"
	"time"
	"fmt"
)

type Movie struct {
	ID     string  `json:"ID"`
	Title  string  `json:"Title"`
	Plot   string  `json:"Plot"`
	Rating float64 `json:"Rating"`
	Year   string  `json:"Year"`
}

// GenerateULID generates a ULID and returns it as a formatted string like 0001, 0002, ...
func GenerateULID() string {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	ulid := ulid.MustNew(ulid.Timestamp(t), entropy)

	// Extract the first 4 characters from the ULID and convert to an integer
	// The first 4 characters should be numeric
	ulidPrefix := ulid.String()[:4]
	var formattedID string

	// Try to parse the first 4 characters of the ULID as an integer
	// You can use this portion to give you an incrementing-like ID
	if num, err := fmt.Sscanf(ulidPrefix, "%d", &formattedID); err == nil {
		// Format as a zero-padded string (e.g., 0001, 0002)
		formattedID = fmt.Sprintf("%04d", num)
	}

	return formattedID
}