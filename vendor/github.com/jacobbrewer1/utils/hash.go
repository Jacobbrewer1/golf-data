package utils

import (
	"crypto/sha1"
	"fmt"
)

func Sha256(data []byte) string {
	hasher := sha1.New()
	hasher.Write(data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
