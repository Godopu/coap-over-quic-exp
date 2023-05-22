package compressedcoap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/udp"
)

var objId = 0

func getObjectID() int {
	objId++
	return objId
}

// coap handler
func (cp *_CompressedCoAPClientProxy) defaultHandler(w mux.ResponseWriter, r *mux.Message) {
	_, _, err := r.Options.Find(message.URIQuery)
	if err != nil {
		if strings.Compare(err.Error(), "option not found") != 0 {
			return
		}
		sm, ok := cp.managers[getHashValue([]byte(w.Client().RemoteAddr().String()))]
		if !ok {
			cp.initialHandler(w, r)
		} else {
			forwardingHandler(w, r, sm)
		}
	} else {
		w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("hello client")))
	}

}

var managerMutex sync.Mutex

// if coap message is general reqeust
func (cp *_CompressedCoAPClientProxy) initialHandler(w mux.ResponseWriter, r *mux.Message) {
	if r.Code == codes.POST {
		if cp.conn == nil {
			panic(errors.New("conn is null"))
		}

		stream, err := cp.conn.OpenStreamSync(context.Background())
		if err != nil {
			w.SetResponse(
				codes.BadRequest,
				message.TextPlain,
				bytes.NewReader([]byte("failed to make stream")),
			)
			return
		}

		co, err := udp.Dial(w.Client().RemoteAddr().String())
		if err != nil {
			w.SetResponse(
				codes.BadRequest,
				message.TextPlain,
				bytes.NewReader([]byte("device do not open coap server")),
			)
			return
		}

		sm, err := NewStreamManager(
			co,
			stream,
		)

		if err != nil {
			return
		}

		managerMutex.Lock()
		cp.managers[getHashValue([]byte(w.Client().RemoteAddr().String()))] = sm
		managerMutex.Unlock()

		sm.DefaultHandleFunc(func(m *message.Message) {
			m.Context = context.Background()
			resp, err := co.Client().Do(m)
			if err != nil {
				return
			}

			fmt.Println("resp:", resp)

			sm.Queue(resp)
		})

		sm.CloseFunc(func() {
			log.Println("sm is closed")
			delete(cp.managers, getHashValue([]byte(w.Client().RemoteAddr().String())))
		})

		err = sm.Run()

		tkn, err := message.GetToken()
		if err != nil {
			sm.Close()
			w.SetResponse(
				codes.BadRequest,
				message.TextPlain,
				bytes.NewReader([]byte("something went wrong when make token")),
			)
		}

		location := []byte(fmt.Sprintf("/75002/%d", getObjectID()))
		initialMsg := &message.Message{
			Token:   tkn,
			Code:    codes.POST,
			Context: context.Background(),
			Body:    bytes.NewReader(location),
		}

		sm.Queue(initialMsg)

		if err != nil {
			sm.Close()
			w.SetResponse(
				codes.BadRequest,
				message.TextPlain,
				bytes.NewReader([]byte("failed to make stream manager")),
			)
			return
		}

		w.SetResponse(
			codes.Created,
			message.TextPlain,
			nil,
			message.Option{
				ID:    message.LocationPath,
				Value: location,
			},
		)

	} else {
		w.SetResponse(
			codes.BadRequest,
			message.TextPlain,
			bytes.NewReader([]byte("you should perform registration procedure")),
		)
	}
}
