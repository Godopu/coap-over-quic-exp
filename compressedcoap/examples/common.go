package examples

import (
	"net"
	"scenarios/compressedcoap"
	"scenarios/mock"
)

func newEntities() (mock.MockServerI, mock.MockClientI, compressedcoap.ProxyI, compressedcoap.ProxyI, error) {
	ms := mock.NewMockServer(":5683")
	sp, err := compressedcoap.NewServerProxy(
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 8080,
		},
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 12126,
		},
		"[::1]:5683",
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cp := compressedcoap.NewClientProxy(
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 9000,
		},
		"[::1]:5683",
	)

	mc, err := mock.NewMockClient(0, 600)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return ms, mc, sp, cp, nil
}

func newEntitiesWithMultipleMC(n int) (mock.MockServerI, []mock.MockClientI, compressedcoap.ProxyI, compressedcoap.ProxyI, error) {
	ms := mock.NewMockServer(":5683")
	sp, err := compressedcoap.NewServerProxy(
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 8080,
		},
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 12126,
		},
		"[::1]:5683",
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cp := compressedcoap.NewClientProxy(
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 9000,
		},
		"[::1]:5683",
	)

	mcList := make([]mock.MockClientI, n)
	for i := 0; i < n; i++ {
		mc, err := mock.NewMockClient(0, 600)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		mcList[i] = mc
	}

	return ms, mcList, sp, cp, nil
}
