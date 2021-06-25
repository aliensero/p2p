package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ma "github.com/multiformats/go-multiaddr"
)

var gh host.Host
var gdhts []*dht.IpfsDHT
var pis []peer.AddrInfo

var tenPeerPriKey = []string{
	"08011240a7fb822435683d724512d13a5fc6d76b3c27f7f87d8c2ee9bc1284a91ffce1555e86a7c0f9c93fe3af485aca06d4c4fabe27c34bb869e097e832608a51781b94",
	"08011240ee2658c574da9dddefe4357518f43d2d2eb822af2676ba1b279ecc3c2835999fb0190d8e79d36c122a5bd1959f4c7fc015ab5c651969ddd514a3e9cb8740354e",
	"08011240a0f93e7a805a9dcc656de5ea33d06f7bd8df1f9a5d6a55839ad736e34507c1c9283728a55b2d3a23b48c3171895bc726941ed3638051967e7cc2d9e22f15d808",
	"08011240747be9dd3ec6fb3a6784195cd922b7ebb7f127c47f9c5127741d106126297f3e7a5aa6f171aa410169f143fe3480223e31a10c6cc4d1c62fdf0b929bfb82bc31",
	"080112405e8bd97127518e92a76036e551b6322e975465ed21507c51262381efde3461d0da913140146922cea168f32bec234807f29fde5a622cca6945f03b644237a0da",
	"08011240d5121ccc3b087db5ec7b32dee185f3c17a96e35bc20ec239515f5b24a6fd9e6d23a1242edadd6a12a19dc7c2cbfd429ccc0a2f3cfbc97d0ed59e3045fad371db",
	"0801124077b6bed26652814d53fbb1b17bb7c562613cb7052d45c69d31c5ac1282f21dfe56c19973b34790b68bf24fa454c818eb962fb9c1b066d936f582748da7d36c45",
	"080112405dcdb43369e2cd57ebe0ac56aef69c500909c99d18ff3849f2df81164f45c70559912ae2d483b911314721232cf7e0450c3a2709b307445cb858dc20426fcffc",
	"080112405d51575192c81bba74adc9dbde113a3b24d38e0bef6735dcbfcd96be2caf7758836ca881b56d80a591d92d2aca5f3935e903e74c066ab7433facd5eba6b68d17",
	"08011240c34142b23b8a6fcd1b5c9f0fc7d9fb7e8bc4a92ba4fa3f34cfb37325bc547691d80ef35d3dcef0402ec3b817cae2645d58dbcad83a07b7843db8c3fa338646c6",
}

var bp = 3456
var protcocolPrefix = "/alien/kad/test"
var topic = "/alien/test"
var announceAddr = "47.115.164.95"

func setupDHT() {
	ctx := context.TODO()
	baseOpts := []dht.Option{
		dht.ProtocolPrefix(protocol.ID(protcocolPrefix)),
		dht.Mode(dht.ModeServer),
	}
	for i := 0; i < 5; i++ {
		tbp := bp + i
		addrFactor := func(allAddrs []ma.Multiaddr) []ma.Multiaddr {
			maddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", announceAddr, tbp))
			if err != nil {
				log.Println(err)
				return nil
			}
			return []ma.Multiaddr{maddr}
		}
		buf, err := hex.DecodeString(tenPeerPriKey[i])
		if err != nil {
			log.Println(err)
			continue
		}
		pk, err := crypto.UnmarshalPrivateKey(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		h, err := libp2p.New(ctx, libp2p.AddrsFactory(addrFactor), libp2p.Identity(pk))
		if err != nil {
			log.Println(err)
			continue
		}

		maddr, err := ma.NewMultiaddr(fmt.Sprintf("%s%d", "/ip4/0.0.0.0/tcp/", tbp))
		if err != nil {
			log.Println(err)
			continue
		}
		err = h.Network().Listen(maddr)
		if err != nil {
			log.Println(err)
			continue
		}

		d, err := dht.New(ctx, h, baseOpts...)
		if err != nil {
			log.Println(err)
			continue
		}
		err = d.Bootstrap(ctx)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(h.Addrs(), " ", h.ID(), " ", d.Host().Network().Peers(), " ", d.RoutingTable().Size())
		if gh == nil {
			gh = h
		}
		pi := peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
		pis = append(pis, pi)
		gdhts = append(gdhts, d)
	}
}

func main() {

	setupDHT()
	for _, pi := range pis[1:] {
		log.Printf("peer addr info %v\n", pi)
		err := gh.Connect(context.TODO(), pi)
		if err != nil {
			log.Println(err)
		}
	}

	ps, err := pubsub.NewGossipSub(context.TODO(), gh)
	if err != nil {
		panic(err)
	}

	t, err := ps.Join(topic)
	if err != nil {
		log.Fatal(err)
	}

	sub, err := t.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			got, err := sub.Next(context.TODO())
			if err != nil {
				log.Fatal(err)
			}
			log.Println("mesage ", string(got.Data))
		}
	}()

	for {
		gdhts[1].RoutingTable().Print()
		time.Sleep(10 * time.Second)
	}
}
