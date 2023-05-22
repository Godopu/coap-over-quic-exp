package notifier

import "github.com/plgd-dev/go-coap/v2/message"

type ISubscriber interface {
	Handle(e IEvent)
	Type() int
	ID() string
	Token() string
}

type ChanSubscriber struct {
	_id      string
	_token   string
	_type    int
	_channel chan<- *message.Message
}

func (rs *ChanSubscriber) ID() string {
	return rs._id
}

func (rs *ChanSubscriber) Token() string {
	return rs._token
}

func (rs *ChanSubscriber) Type() int {
	return rs._type
}

func (rs *ChanSubscriber) Handle(e IEvent) {
	rs._channel <- e.Body().(*message.Message)
}

func NewChanSubscriber(_id, _token string, _type int, _channel chan<- *message.Message) ISubscriber {
	return &ChanSubscriber{_id: _id, _token: _token, _type: _type, _channel: _channel}
}
