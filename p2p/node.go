package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"gitlab.com/thesepehrm/p2p-sync/common"
)

const (
	ProtocolPath = "/main/1.0.0"
)

type Node struct {
	host host.Host

	privateKey crypto.PrivKey
	source     multiaddr.Multiaddr

	data map[common.Hash]string
}

func NewNode(sourcePort int) *Node {

	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", sourcePort))
	if err != nil {
		log.Panic(err)
	}

	privateKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	if err != nil {
		log.Panic(err)
	}

	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(privateKey),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
	)

	if err != nil {
		log.Panic(err)
	}

	n := &Node{
		host:       host,
		privateKey: privateKey,
		source:     sourceMultiAddr,

		data: make(map[common.Hash]string),
	}

	return n
}

func (n *Node) Start() {

	var port string
	for _, la := range n.host.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			fmt.Printf("Run './chat -d /ip4/127.0.0.1/tcp/%v/p2p/%s' on another console.\n", port, n.host.ID().Pretty())
		}
	}

	if port == "" {
		panic("was not able to find actual local port")
	}

	fmt.Println("You can replace 127.0.0.1 with public IP as well.")
	fmt.Printf("\nWaiting for incoming connection\n\n")

	n.host.SetStreamHandler(ProtocolPath, handleStream)

}

func (n *Node) Connect(dest string) error {

	destAddr := common.StringToMultiAddr(dest)

	fmt.Println("This node's multiaddresses:")
	for _, la := range n.host.Addrs() {
		fmt.Printf(" - %v\n", la)
	}
	fmt.Println()

	destInfo, err := peer.AddrInfoFromP2pAddr(destAddr)
	if err != nil {
		return err
	}
	n.host.Peerstore().AddAddrs(destInfo.ID, destInfo.Addrs, peerstore.PermanentAddrTTL)

	stream, err := n.host.NewStream(context.Background(), destInfo.ID, ProtocolPath)
	if err != nil {
		log.Panic(err)
	}

	handleStream(stream)

	return nil
}

func handleStream(s network.Stream) {
	fmt.Println("got new stream")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}

}

func writeData(rw *bufio.ReadWriter) {

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}

}
