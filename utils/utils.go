package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func ShortenURLHash(longURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(longURL))
	hashValue := hex.EncodeToString(hasher.Sum(nil))
	return hashValue[:5] // Take the first 5 characters of the hash
}
