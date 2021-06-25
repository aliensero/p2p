package pubsub

import (
	"context"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	ldht "github.com/aliensero/p2p-modules/modules/dht"
	lhost "github.com/aliensero/p2p-modules/modules/host"
	"github.com/libp2p/go-libp2p-core/host"
)

var log = logging.Logger("pubsub")

var topic = "/alien/test"

func init() {
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		logging.SetLogLevel("pubsub", v)
	} else {
		logging.SetLogLevel("pubsub", "INFO")
	}
}

type PubSub struct {
	*pubsub.PubSub
	host.Host
	*dht.IpfsDHT
}

func (p *PubSub) Close() {
	p.IpfsDHT.Close()
	p.Host.Close()
}

func DefaultPubsub() (*PubSub, error) {
	ctx := context.TODO()

	h, err := lhost.DefaultHost()
	if err != nil {
		return nil, err
	}
	d, err := ldht.DefaultDHT(dht.ModeServer, h)
	if err != nil {
		return nil, err
	}
	go func() {
		errCnt := 0
		for {
			err := <-d.RefreshRoutingTable()
			if err != nil {
				log.Error(errCnt, err)
				if errCnt == 10 {
					if th, ok := h.(lhost.Host); ok {
						err = th.ConnectBootStrap()
						if err != nil {
							log.Error(err)
						}
					}
					errCnt = 0
				}
			}
			log.Debug(d.RoutingTable().ListPeers())
			time.Sleep(30 * time.Second)
		}
	}()

	ps, err := pubsub.NewGossipSub(ctx, h, options...)
	if err != nil {
		return nil, err
	}

	lsub := &PubSub{
		ps,
		h,
		d.IpfsDHT,
	}

	return lsub, nil
}
