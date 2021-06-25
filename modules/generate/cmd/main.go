package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func main() {
	for i := 0; i < 10; i++ {
		pk, _, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			continue
		}
		buf, err := crypto.MarshalPrivateKey(pk)
		if err != nil {
			continue
		}
		fmt.Println(len(buf))
		fmt.Println(hex.EncodeToString(buf))
	}
}
