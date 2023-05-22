package compressedcoap

import (
	"context"
	"io"
	"scenarios/compressedcoap/converter"
	"scenarios/compressedcoap/wire"
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

// todo
// - add receive on transmitToSP
// - add event listener

// device <- (ip, port) -> cp <-stream-> sp
// need id (ip and port) + stream
type streamManager struct {
	devConn *client.ClientConn    // this value is derieved from IP and port of device
	enc     converter.FrameWriter // to sp
	dec     converter.Decoder
	stream  io.ReadWriter

	hijacker          map[uint64]chan *message.Message
	hijackerMutex     sync.Mutex
	mapper            tokenMapper
	msgQueue          chan *message.Message
	defaultHandleFunc func(*message.Message)
	initialFunc       func(*message.Message) error
	closeFunc         func()
	cancelFunc        context.CancelFunc
}

type StreamManager interface {
	Queue(msg *message.Message)
	Hijack(token string) *message.Message
	DefaultHandleFunc(func(*message.Message))
	InitialFunc(func(*message.Message) error)
	CloseFunc(func())
	Run() error
	// Send(msg *message.Message, upt uint8) error
	// Recv() (*message.Message, uint8, error)
	Close()
}

func (dlv *streamManager) Close() {
	dlv.devConn.Close()
	qstream, ok := dlv.stream.(quic.Stream)
	if ok {
		qstream.Close()
	}
	if dlv.cancelFunc != nil {
		dlv.cancelFunc()
	}

	for _, v := range dlv.hijacker {
		close(v)
	}
}

func (dlv *streamManager) Queue(msg *message.Message) {
	dlv.msgQueue <- msg
}

func (dlv *streamManager) addHijacker(token string, respCh chan *message.Message) {
	dlv.hijackerMutex.Lock()
	defer dlv.hijackerMutex.Unlock()
	dlv.hijacker[getHashValue([]byte(token))] = respCh
}

func (dlv *streamManager) rmHijacker(token string) {
	dlv.hijackerMutex.Lock()
	defer dlv.hijackerMutex.Unlock()
	delete(dlv.hijacker, getHashValue([]byte(token)))
}

func (dlv *streamManager) getHijacker(token string) chan *message.Message {
	dlv.hijackerMutex.Lock()
	defer dlv.hijackerMutex.Unlock()
	ch, ok := dlv.hijacker[getHashValue([]byte(token))]
	if !ok {
		return nil
	}

	return ch
}

func (dlv *streamManager) Hijack(token string) *message.Message {
	respCh := make(chan *message.Message)
	dlv.addHijacker(token, respCh)

	msg, ok := <-respCh
	if !ok {
		return nil
	}

	dlv.rmHijacker(token)
	return msg
}

func (dlv *streamManager) DefaultHandleFunc(defaultHandleFunc func(*message.Message)) {
	dlv.defaultHandleFunc = defaultHandleFunc
}

func (dlv *streamManager) InitialFunc(initialFunc func(*message.Message) error) {
	dlv.initialFunc = initialFunc
}
func (dlv *streamManager) CloseFunc(closeFunc func()) {
	dlv.closeFunc = closeFunc
}

func (dlv *streamManager) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	dlv.cancelFunc = cancel

	if dlv.initialFunc != nil {
		msg, _, err := dlv.recv()
		if err != nil {
			return err
		}

		err = dlv.initialFunc(msg)
		if err != nil {
			return err
		}
	}
	go dlv.routine(ctx)

	return nil
}

func (dlv *streamManager) routine(ctx context.Context) {
	go func() {
		// perform listen recv
		for {
			msg, _, err := dlv.recv()
			if err != nil {
				return
			}

			msgCh := dlv.getHijacker(msg.Token.String())
			if msgCh != nil {
				msgCh <- msg
				continue
			}

			// handle request
			dlv.defaultHandleFunc(msg)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-dlv.msgQueue:
			if !ok {
				return
			}
			err := dlv.send(req, wire.UDP)
			if err != nil {
				return
			}
		}
	}
}

func (dlv *streamManager) send(msg *message.Message, upt uint8) error {
	// if message is request
	if msg.Code&0xE0 == 0 {
		cmpTkn, seq, err := dlv.enc.Request(msg, upt)

		if err != nil {
			return err
		}
		dlv.mapper.pushEncodedMsgTkn(msg.Token, cmpTkn, seq)

		return nil
	}

	// if message is response
	tkn, seq, err := dlv.mapper.popDecodedMsgTkn(msg.Token)
	if err != nil {
		return err
	}

	msg.Token = tkn
	err = dlv.enc.Response(msg, seq)
	if err != nil {
		return err
	}

	return nil
}

func (dlv *streamManager) recv() (*message.Message, uint8, error) {
	var decodedMsg message.Message
	var upt uint8

	seq, err := dlv.dec.Decode(&decodedMsg, &upt)
	if err != nil {
		return nil, 0, err
	}

	if decodedMsg.Code&0xE0 == 0 {
		// handle request message
		newTkn, err := message.GetToken()
		if err != nil {
			return nil, 0, err
		}
		err = dlv.mapper.pushDecodedMsgTkn(newTkn, decodedMsg.Token, seq)
		if err != nil {
			return nil, 0, err
		}

		decodedMsg.Token = newTkn
		return &decodedMsg, 0, err
	}

	// handle response message
	originTkn, err := dlv.mapper.popEncodedMsgTkn(decodedMsg.Token, seq)
	if err != nil {
		return nil, 0, err
	}

	decodedMsg.Token = originTkn
	return &decodedMsg, upt, nil
}

func NewStreamManager(co *client.ClientConn, stream io.ReadWriter) (StreamManager, error) {
	sm := &streamManager{
		devConn:  co,
		enc:      converter.NewWriter(stream),
		dec:      converter.NewDecoder(stream),
		stream:   stream,
		hijacker: map[uint64]chan *message.Message{},
		mapper: tokenMapper{
			encodedMsgTkns: map[uint64][][]byte{},
			decodedMsgTkns: map[uint64]*compressedTknPair{},
		},
		msgQueue: make(chan *message.Message, 2000),
	}

	return sm, nil
}
