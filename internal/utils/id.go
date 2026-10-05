package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a random hexadecimal string identifier.
func NewID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes[:])
}
