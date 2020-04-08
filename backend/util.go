package backend

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

// Computehash returns the hash of its inputs
func Computehash(data ...[]byte) []byte {
	hasher := sha256.New()
	for _, b := range data {
		_, err := hasher.Write(b)
		if err != nil {
			panic(1)
		}
	}
	return hasher.Sum(nil)
}

func makeCENKeyString() string {
	key := make([]byte, 16)
	rand.Read(key)
	encoded := fmt.Sprintf("%x", key)
	return encoded
}
