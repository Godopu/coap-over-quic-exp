package compressedcoap

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"scenarios/mock"
	"testing"
	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/udp"
	"github.com/stretchr/testify/assert"
)

func TestPerformInitialize(t *testing.T) {
	assert := assert.New(t)

	ms, _, sp, cp, err := newEntities()

	go ms.Run()
	go sp.Run()

	assert.NoError(err)

	time.Sleep(time.Millisecond * 50)

	laddr, _, err := cp.(*_CompressedCoAPClientProxy).performInitialize()
	assert.NoError(err)

	co, err := udp.Dial("[::1]:5683")
	assert.NoError(err)

	resp, err := co.Get(context.Background(), "/rd", message.Option{
		ID:    message.URIQuery,
		Value: []byte(fmt.Sprintf("addr=%s", laddr.String())),
	})
	assert.NoError(err)

	assert.Equal(resp.Code(), codes.Content)

	body, err := ioutil.ReadAll(resp.Body())
	assert.NoError(err)
	assert.Equal("client-proxy", string(body))
}

func TestCPRun(t *testing.T) {
	assert := assert.New(t)

	ms, _, sp, cp, err := newEntities()

	go ms.Run()
	go sp.Run()

	assert.NoError(err)

	time.Sleep(time.Millisecond * 50)

	err = cp.Run()
	assert.NoError(err)
}

func newEntities() (mock.MockServerI, mock.MockClientI, ProxyI, ProxyI, error) {
	ms := mock.NewMockServer(":5683")
	sp, err := NewServerProxy(
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

	cp := NewClientProxy(
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
