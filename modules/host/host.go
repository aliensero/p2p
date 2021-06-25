package host

import (
	"context"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("host")
var bootstrapFile = "bootstrap.pi"

type Host struct {
	host.Host
	bootstrapAddrs []peer.AddrInfo
}

func (h *Host) ConnectBootStrap() error {
	ctx := context.TODO()
	var err error
	for _, p := range h.bootstrapAddrs {
		err = h.Connect(ctx, p)
		if err == nil {
			break
		}
	}
	return err
}

var HostOpts = []libp2p.Option{
	libp2p.ConnectionManager(connmgr.NewConnManager(50, 200, 20*time.Second)),
	libp2p.NoListenAddrs,
}

func DefaultHost() (host.Host, error) {
	ctx := context.TODO()

	h, err := libp2p.New(ctx, HostOpts...)
	if err != nil {
		return nil, err
	}
	pis, err := getBootstrapAddrInfos()
	if err != nil {
		return nil, err
	}

	lh := &Host{
		h,
		pis,
	}
	err = lh.ConnectBootStrap()
	if err != nil {
		return nil, err
	}
	return lh, nil
}

func getBootstrapAddrInfos() ([]peer.AddrInfo, error) {
	b := rice.MustFindBox("../generate/bootstrap")
	maddrstrs := b.MustString(bootstrapFile)
	maddrArr := strings.Split(strings.TrimSpace(maddrstrs), "\n")

	var mas []ma.Multiaddr
	for _, maa := range maddrArr {
		ma, err := ma.NewMultiaddr(maa)
		if err != nil {
			log.Errorf("NewMutltiaddr %s error %v", maa, err)
			continue
		}
		mas = append(mas, ma)
	}
	pis, err := peer.AddrInfosFromP2pAddrs(mas...)
	if err != nil {
		return nil, err
	}
	return pis, nil
}
