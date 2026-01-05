package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a random hexadecimal string identifier.
func NewID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
