package notifier

import "github.com/plgd-dev/go-coap/v2/message"

type IEvent interface {
	Token() string
	Title() string
	Body() interface{}
}

// Default implementation of IEvent
type RecvRespEvent struct {
	_title string
	_body  *message.Message
	_token string
}

func (e *RecvRespEvent) Token() string {
	return e._token
}

func (e *RecvRespEvent) Title() string {
	return e._title
}

func (e *RecvRespEvent) Body() interface{} {
	return e._body
}

func NewRecvRespEvent(_title, _token string, _body *message.Message) IEvent {
	return &RecvRespEvent{
		_title,
		_body,
		_token,
	}
}
