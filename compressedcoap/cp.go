package compressedcoap

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	gonet "net"
	"os"
	"scenarios/compressedcoap/notifier"

	"github.com/lucas-clemente/quic-go"
	"github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/udp"
)

// this address will be received from iot server
var serverProxyAddr = "localhost:8080"
var notifyBox = notifier.NewNotiManager()

type _CompressedCoAPClientProxy struct {
	coapBindAddr *net.UDPAddr
	serverAddr   string
	managers     map[uint64]StreamManager
	conn         quic.Connection
}

func NewClientProxy(coapBindAddr *net.UDPAddr, serverAddr string) ProxyI {

	// init

	return &_CompressedCoAPClientProxy{
		coapBindAddr,
		serverAddr,
		map[uint64]StreamManager{},
		nil,
	}
}

func (cp *_CompressedCoAPClientProxy) runCoAPServer() error {
	r := mux.NewRouter()
	r.DefaultHandleFunc(cp.defaultHandler)

	return coap.ListenAndServe(
		"udp",
		fmt.Sprintf(":%d", cp.coapBindAddr.Port),
		r,
	)
}

// return localAddr, remoteAddr, error
func (cp *_CompressedCoAPClientProxy) performInitialize() (*gonet.UDPAddr, *gonet.UDPAddr, error) {
	co, err := udp.Dial(cp.serverAddr, udp.WithDialer(&net.Dialer{
		Control: reusePort,
	}))
	if err != nil {
		return nil, nil, err
	}

	payload := bytes.NewReader([]byte("/75002/1/0"))

	resp, err := co.Post(
		context.Background(),
		"/rd",
		message.TextPlain,
		payload,
		message.Option{
			ID:    message.URIQuery,
			Value: []byte("ep=client-proxy"),
		},
	)

	if err != nil {
		return nil, nil, err
	}

	if resp.Code() != codes.Created {
		return nil, nil, errors.New("something went wrong on server")
	}

	body, err := ioutil.ReadAll(resp.Body())
	if err != nil {
		return nil, nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp", string(body))
	if err != nil {
		return nil, nil, err
	}

	laddr, ok := co.LocalAddr().(*gonet.UDPAddr)
	if !ok {
		return nil, nil, errors.New("invalid local address")
	}

	return laddr, raddr, nil
}

func (cp *_CompressedCoAPClientProxy) Run() error {

	laddr, raddr, err := cp.performInitialize()
	if err != nil {
		return err
	}

	keylog, err := os.Create("./ssl-key.log")
	if err != nil {
		return err
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		KeyLogWriter:       keylog,
		NextProtos:         []string{"quic-echo-example"},
	}

	lc := gonet.ListenConfig{
		Control: reusePort,
	}

	lp, err := lc.ListenPacket(
		context.Background(),
		"udp",
		fmt.Sprintf(":%d", laddr.Port),
	)

	if err != nil {
		return err
	}

	conn, err := quic.Dial(lp, raddr, raddr.String(), tlsConf, nil)
	if err != nil {
		return err
	}
	cp.conn = conn

	cp.runCoAPServer()

	return nil
}
