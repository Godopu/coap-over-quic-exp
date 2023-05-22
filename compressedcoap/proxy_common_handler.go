package compressedcoap

import (
	"context"

	"github.com/plgd-dev/go-coap/v2/mux"
)

func forwardingHandler(w mux.ResponseWriter, r *mux.Message, sm StreamManager) {

	// upt := wire.UDP
	// if strings.Compare("tcp", w.Client().RemoteAddr().Network()) == 0 {
	// 	upt = wire.TCP
	// }

	sm.Queue(r.Message)
	resp := sm.Hijack(r.Token.String())
	resp.Context = context.Background()
	w.Client().WriteMessage(resp)

	// w.SetResponse(resp.Code, message.TextPlain, resp.Body, resp.Options...)
}
