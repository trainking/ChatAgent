package websocket

import (
	"fmt"
	"math/rand"
)

func GeneratePubsubToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
