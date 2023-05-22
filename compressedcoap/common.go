package compressedcoap

import (
	"hash/crc64"
	"syscall"

	"golang.org/x/sys/unix"
)

func getHashValue(b []byte) uint64 {
	return crc64.Checksum(b, crc64.MakeTable(crc64.ISO))
}

func reusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
}
