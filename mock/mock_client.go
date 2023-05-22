package mock

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	coapnet "github.com/plgd-dev/go-coap/v2/net"
	"github.com/plgd-dev/go-coap/v2/udp"
	"github.com/plgd-dev/go-coap/v2/udp/client"
	"golang.org/x/sys/unix"
)

type _mockClient struct {
	mean float64
	id   string
	co   *client.ClientConn
}

type MockClientI interface {
	Connect(serverAdr string) (string, error)
	Client() *client.Client
}

func (mc *_mockClient) Client() *client.Client {
	if mc.co == nil {
		return nil
	}

	return mc.co.Client()
}

func reusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
}

func (mc *_mockClient) Connect(serverAdr string) (string, error) {
	co, err := udp.Dial(serverAdr, udp.WithDialer(&net.Dialer{
		Control: reusePort,
	}))
	if err != nil {
		return "", err
	}

	resp, err := co.Client().Post(
		context.Background(),
		"/rd",
		message.TextPlain,
		nil,
	)
	if err != nil {
		return "", err
	}

	if resp.Code != codes.Created {
		return "", errors.New("something went wrong on server")
	}

	first, _, err := resp.Options.Find(message.LocationPath)
	if err != nil {
		return "", err
	}

	location := string(resp.Options[first].Value)

	mc.id = location
	mc.co = co
	go runServer(co.LocalAddr().(*net.UDPAddr))
	return location, nil
}

func runServer(addr *net.UDPAddr) error {
	lc := net.ListenConfig{
		Control: reusePort,
	}

	lp, err := lc.ListenPacket(
		context.Background(),
		"udp",
		fmt.Sprintf(":%d", addr.Port),
	)

	if err != nil {
		return err
	}

	conn := lp.(*net.UDPConn)

	rt := mux.NewRouter()
	rt.DefaultHandleFunc(func(w mux.ResponseWriter, r *mux.Message) {
		w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("hello server")))
	})

	s := udp.NewServer(udp.WithMux(rt))
	return s.Serve(coapnet.NewUDPConn("udp", conn))
}

func NewMockClient(id int, mean float64) (MockClientI, error) {
	return &_mockClient{mean, "", nil}, nil
}
