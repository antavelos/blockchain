package blockchain

import (
	"crypto/sha512"
	"fmt"

	"github.com/google/uuid"
)

func newUuid() string {
	return fmt.Sprintf("%v", uuid.New())
}

func hash(hashable []byte) []byte {
	hash := sha512.New()

	hash.Write(hashable)

	return hash.Sum(nil)
}
