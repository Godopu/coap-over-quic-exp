package main

import (
	"flag"
	"net"
	"scenarios/compressedcoap"
)

func main() {
	qbind := flag.String("qbind", ":8080", "coap server bind address")
	cbind := flag.String("cbind", ":12126", "quic server bind address")
	flag.Parse()
	serverAdr := flag.Arg(0)

	cbindAddr, err := net.ResolveUDPAddr("udp", *cbind)
	if err != nil {
		panic(err)
	}

	qbindAddr, err := net.ResolveUDPAddr("udp", *qbind)
	if err != nil {
		panic(err)
	}

	if qbindAddr.IP == nil {
		qbindAddr.IP = net.ParseIP("0.0.0.0")
	}
	sp, err := compressedcoap.NewServerProxy(
		qbindAddr,
		cbindAddr,
		serverAdr,
	)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	sp.Run()

}
