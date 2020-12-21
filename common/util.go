package common

import (
	"log"

	"github.com/multiformats/go-multiaddr"
)

func StringToMultiAddr(addr string) multiaddr.Multiaddr {
	multiaddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Panic(err)
	}
	return multiaddr
}
