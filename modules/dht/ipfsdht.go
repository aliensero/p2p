package route

import (
	"context"

	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	gen "github.com/aliensero/p2p-modules/modules/generate"
	lhost "github.com/aliensero/p2p-modules/modules/host"
	"github.com/libp2p/go-libp2p-core/host"
)

type DHT struct {
	host.Host
	*dht.IpfsDHT
}

func (d *DHT) Close() {
	d.IpfsDHT.Close()
	d.Host.Close()
}

func DefaultDHT(mode dht.ModeOpt, h ...host.Host) (*DHT, error) {
	ctx := context.TODO()

	opts := []dht.Option{
		dht.Mode(mode),
		dht.ProtocolPrefix(protocol.ID(gen.KadPrefix)),
	}
	var host host.Host
	if len(h) == 0 {
		var err error
		host, err = lhost.DefaultHost()
		if err != nil {
			return nil, err
		}
	} else {
		host = h[0]
	}
	d, err := dht.New(
		ctx, host, opts...,
	)
	if err != nil {
		return nil, err
	}

	ld := &DHT{
		host,
		d,
	}

	return ld, nil
}
