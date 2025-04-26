package models

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

type Movie struct {
	ID     string `json:"ID"`
	Title  string `json:"Title"`
	Plot   string `json:"Plot"`
	Rating int    `json:"Rating"` // int
	Year   string `json:"Year"`
}

// ulid
func GenerateULID() string {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
