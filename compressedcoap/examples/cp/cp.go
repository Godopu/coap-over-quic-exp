package main

import (
	"flag"
	"net"
	"scenarios/compressedcoap"
)

func main() {
	b := flag.String("bind", ":9000", "bind address")
	flag.Parse()
	serverAdr := flag.Arg(0)

	bindAddr, err := net.ResolveUDPAddr("udp", *b)
	if err != nil {
		panic(err)
	}

	cp := compressedcoap.NewClientProxy(
		bindAddr,
		serverAdr,
	)

	err = cp.Run()
	if err != nil {
		panic(err)
	}
}
