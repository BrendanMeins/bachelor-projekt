package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BrendanMeins/bachelor-projekt/consensus"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	disc "github.com/libp2p/go-libp2p-discovery"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
)

type discoveryNotifee struct {
	h   host.Host
	ctx context.Context
}

//Handler Funktion, die dafür sorgt, dass neue Peers dem Adressbuchg
//hinzugefügt werden
func (m *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if m.h.Network().Connectedness(pi.ID) != network.Connected {
		fmt.Printf("Found %s!\n", pi.ID.Pretty())
		m.h.Connect(m.ctx, pi)
	}
}

var Message *string

//In der Main Methode wird der Host von Libp2p aufgesetzt. Und die
//Schleife gestartet, die den User Input aus dem Terminal erwartet.
func main() {
	bootstrapNodeAddress := flag.String("b", "", "Gibt den Node eine Adresse zu einem Bootsrap Node")
	Message = flag.String("m", "", "Message zum signieren")
	flag.Parse()
	transports := libp2p.ChainOptions(libp2p.Transport(tcp.NewTCPTransport))
	listenAddrs := libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/tcp/0",
	)
	ctx := context.Background()
	var dht *kaddht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dht, err = kaddht.New(ctx, h)
		return dht, err
	}
	routing := libp2p.Routing(newDHT)

	h, err := libp2p.New(
		transports,
		listenAddrs,
		routing,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(h.Addrs()[0].String() + "/p2p/" + h.ID().Pretty())
	notifee := &discoveryNotifee{h: h, ctx: ctx}
	mdns := mdns.NewMdnsService(h, "", notifee)
	consensus.Init(h)
	if *bootstrapNodeAddress != "" {
		targetAddr, err := multiaddr.NewMultiaddr(*bootstrapNodeAddress)
		if err != nil {
			panic(err)
		}
		targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
		if err != nil {
			panic(err)
		}
		h.Peerstore().AddAddrs(targetInfo.ID, targetInfo.Addrs, peerstore.PermanentAddrTTL)
		err = h.Connect(ctx, *targetInfo)

	}
	if err := mdns.Start(); err != nil {
		panic(err)
	}

	err = dht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}

	routingDiscovery := disc.NewRoutingDiscovery(dht)
	disc.Advertise(ctx, routingDiscovery, string(consensus.KeyGenProtocolId))
	peers, err := disc.FindPeers(ctx, routingDiscovery, string(consensus.KeyGenProtocolId))
	if err != nil {
		panic(err)
	}
	for _, peer := range peers {
		notifee.HandlePeerFound(peer)
	}

	cli()
}

func createHost() host.Host {
	transports := libp2p.ChainOptions(libp2p.Transport(tcp.NewTCPTransport))
	listenAddrs := libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/tcp/0",
	)
	ctx := context.Background()
	var dht *kaddht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dht, err = kaddht.New(ctx, h)
		return dht, err
	}
	routing := libp2p.Routing(newDHT)

	h, err := libp2p.New(
		transports,
		listenAddrs,
		routing,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(h.Addrs()[0].String() + "/p2p/" + h.ID().Pretty())
	return h
}

func cli() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		text := scanner.Text()
		if strings.Contains(text, "keygen") {
			consensus.StartKeyGen()
		} else if strings.Contains(text, "sign") {
			msg := strings.Split(text, " ")[1]
			initmsg := consensus.GetInitMessage(msg)
			fmt.Println("init sign with params: ", initmsg)
			consensus.StartSign(*initmsg)
		} else if strings.Contains(text, "hosts") {
			for _, peer := range consensus.Host.Peerstore().Peers() {
				fmt.Printf("peer id : %s ||  partyId: %v\n", peer.String(), consensus.GetPartyId(peer))
			}
		} else {

			fmt.Println("Unknown command")
		}

	}

}
