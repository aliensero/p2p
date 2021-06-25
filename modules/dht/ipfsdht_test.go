package route

import (
	"log"
	"testing"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func TestIpfsDHT(t *testing.T) {

	d, err := DefaultDHT(dht.ModeServer)
	if err != nil {
		t.Error(err)
	}

	for {
		err := <-d.RefreshRoutingTable()
		log.Println("refresh ", err)
		d.RoutingTable().Print()
		time.Sleep(3 * time.Second)
	}
}
