package compressedcoap

import (
	"bytes"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

// implementing
func checkObjectValidation() bool {
	return true
}

func (sp *_CompressedCoAPServerProxy) defaultCoAPHandler(w mux.ResponseWriter, r *mux.Message) {

	first, _, err := r.Options.Find(message.ProxyURI)
	if err != nil {
		w.SetResponse(
			codes.BadRequest,
			message.TextPlain,
			bytes.NewReader([]byte(err.Error())),
		)
	}

	proxyUri := string(r.Options[first].Value)
	cm, ok := sp.managers[proxyUri]
	if !ok {
		w.SetResponse(
			codes.BadRequest,
			message.TextPlain,
			bytes.NewReader([]byte("invalid proxy uri")),
		)
	}

	path, err := r.Options.Path()
	if err != nil {
		w.SetResponse(
			codes.BadRequest,
			message.TextPlain,
			bytes.NewReader([]byte(err.Error())),
		)
	}

	sm := cm.getStreamManager(path)
	if sm == nil {
		w.SetResponse(
			codes.BadRequest,
			message.TextPlain,
			bytes.NewReader([]byte("invalid uri path")),
		)
	}

	forwardingHandler(w, r, sm)
}
