package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/vmihailenco/msgpack/v5"
	"gitlab.com/thesepehrm/p2p-sync/common"
)

const (
	ProtocolPath = "/main/1.0.0"
)

type Node struct {
	host host.Host

	privateKey crypto.PrivKey
	source     multiaddr.Multiaddr

	rw *bufio.ReadWriter

	connected bool

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

		connected: false,

		data: make(map[common.Hash]string),
	}

	return n
}

func (n *Node) Start() {

	var port string
	for _, la := range n.host.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			fmt.Printf("Run \n\tgo run chat.go -d /ip4/127.0.0.1/tcp/%v/p2p/%s\non another console.\n", port, n.host.ID().Pretty())
		}
	}

	if port == "" {
		panic("was not able to find actual local port")
	}

	fmt.Println("You can replace 127.0.0.1 with public IP as well.")
	fmt.Printf("\nWaiting for incoming connection\n\n")

	n.host.SetStreamHandler(ProtocolPath, n.handleStream)

}

func (n *Node) Connect(dest string) error {

	destAddr := common.StringToMultiAddr(dest)

	destInfo, err := peer.AddrInfoFromP2pAddr(destAddr)
	if err != nil {
		return err
	}
	n.host.Peerstore().AddAddrs(destInfo.ID, destInfo.Addrs, peerstore.PermanentAddrTTL)

	stream, err := n.host.NewStream(context.Background(), destInfo.ID, ProtocolPath)
	if err != nil {
		log.Panic(err)
	}

	n.handleStream(stream)

	return nil
}

func (n *Node) handleStream(s network.Stream) {
	fmt.Println("+ New Node Connected!")
	n.rw = bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go n.receive()
	go n.runConsole()
}

func (n *Node) runConsole() {

	stdReader := bufio.NewReader(os.Stdin)

	fmt.Println("Commands:")
	fmt.Println("\t- ping")
	fmt.Println("\t- newdata <data>")
	for {
		fullCommand, _ := stdReader.ReadString('\n')
		fullCommand = strings.TrimRight(fullCommand, "\n\r")

		commandsArgs := strings.Split(fullCommand, " ")

		command := strings.ToLower(commandsArgs[0])
		body := strings.Join(commandsArgs[1:], " ")

		switch command {

		case "ping":

			n.send(&PingPacket{false, time.Now().UnixNano()})

		case "status":
			n.send(&StatusPacket{
				NodeAddress:   "Hello",
				KnownNodesNum: 2,
			})
		case "newdata":
			if len(body) == 0 {
				fmt.Println("Data cannot be empty")
				break
			}

		default:
			fmt.Println("Unknown command: " + command)
		}

	}

}

func (n *Node) receive() {

	for {

		msg, _ := n.rw.ReadByte()

		decoder := msgpack.NewDecoder(n.rw)

		switch Message(msg) {
		case PingMsg:
			var ping PingPacket
			err := decoder.Decode(&ping)
			if err != nil {
				fmt.Println(err)
			}

			pingMs := float64((time.Now().UnixNano() - ping.SentTime)) / 1000000.0

			if !ping.IsReply {
				n.send(&PingPacket{true, time.Now().UnixNano()})

				fmt.Printf("> \x1b[32m%s\x1b[0m (%.02f ms)\n", "requested ping", pingMs)
				fmt.Println()

			} else {
				// Green console colour: 	\x1b[32m
				// Reset console colour: 	\x1b[0m
				fmt.Printf("> \x1b[32m%s\x1b[0m (%.02f ms)\n", "pong", pingMs)
				fmt.Println()
			}

		case StatusMsg:
			var statusPacket StatusPacket
			err := decoder.Decode(&statusPacket)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(statusPacket)

		}

	}
}

func (n *Node) send(data Packet) {

	_ = n.rw.WriteByte(byte(data.Type()))

	encoder := msgpack.NewEncoder(n.rw)
	err := encoder.Encode(data)
	if err != nil {
		fmt.Println(err)
	}

	n.rw.Flush()

}
