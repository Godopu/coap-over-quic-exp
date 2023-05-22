package examples

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/udp"
	"github.com/stretchr/testify/assert"
)

const cnt = 10

func TestTransferReqFromDevToSrv(t *testing.T) {
	assert := assert.New(t)
	ms, mc, sp, cp, err := newEntities()
	assert.NoError(err)

	go ms.Run()
	go func() {
		err := sp.Run()
		assert.NoError(err)
	}()

	time.Sleep(time.Millisecond * 100)

	go func() {
		err := cp.Run()
		assert.NoError(err)
	}()
	assert.NoError(err)

	time.Sleep(time.Millisecond * 100)

	location, err := mc.Connect("localhost:9000")
	assert.NoError(err)

	var wg sync.WaitGroup
	wg.Add(cnt)

	for i := 0; i < cnt; i++ {
		go func() {
			payload := &bytes.Buffer{}
			jsonEncoder := json.NewEncoder(payload)
			err := jsonEncoder.Encode(
				map[string]interface{}{
					"key": "value",
				},
			)
			assert.NoError(err)

			resp, err := mc.Client().Put(
				context.Background(),
				location,
				message.AppJSON,
				bytes.NewReader(payload.Bytes()),
			)
			assert.NoError(err)

			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(err)
			assert.Equal(string(body), "hello client")

			time.Sleep(time.Millisecond * 100)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestTransferReqFromSrvToDev(t *testing.T) {
	assert := assert.New(t)

	ms, mc, sp, cp, err := newEntities()
	assert.NoError(err)

	go ms.Run()
	go func() {
		err := sp.Run()
		assert.NoError(err)
	}()

	time.Sleep(time.Millisecond * 100)

	go func() {
		err := cp.Run()
		assert.NoError(err)
	}()
	assert.NoError(err)

	time.Sleep(time.Millisecond * 100)

	location, err := mc.Connect("localhost:9000")
	assert.NoError(err)
	assert.Equal(location, "/75002/1")

	co, err := udp.Dial("localhost:5683")
	assert.NoError(err)

	resp, err := co.Client().Get(
		context.Background(),
		"/",
	)

	assert.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)

	assert.NoError(err)

	co2, err := udp.Dial("localhost:12126")
	assert.NoError(err)

	for i := 0; i < cnt; i++ {
		resp2, err := co2.Client().Put(
			context.Background(),
			location,
			message.TextPlain,
			bytes.NewReader([]byte("Hello client")),
			message.Option{
				ID:    message.ProxyURI,
				Value: body,
			},
		)
		assert.NoError(err)

		body2, err := ioutil.ReadAll(resp2.Body)
		assert.NoError(err)
		assert.Equal(string(body2), "hello server")
	}

}
