package oauth

import (
	"crypto/rand"
	"encoding/hex"
)

func MustGenerateState() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic("rand read error")
	}
	state := hex.EncodeToString(buf)
	return state
}
