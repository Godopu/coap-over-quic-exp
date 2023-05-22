package mock

import (
	"bytes"
	"log"
	"net"
	"strings"

	"github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

const serverProxyAddr = "10.1.0.1:8080"

type MockServerI interface {
	Run()
}
type _mockServer struct {
	bind   string
	cplist map[string]string
}

func (ms *_mockServer) handle(w mux.ResponseWriter, r *mux.Message) {
	path, err := r.Options.Path()
	if err != nil || len(path) <= 1 {
		for k, _ := range ms.cplist {
			w.SetResponse(
				codes.Content,
				message.TextPlain,
				bytes.NewReader([]byte(k)),
			)
			return
		}
	}

	w.SetResponse(
		codes.Content,
		message.TextPlain,
		bytes.NewReader([]byte("hello client")),
	)
}

func (ms *_mockServer) Run() {
	r := mux.NewRouter()
	r.DefaultHandleFunc(ms.handle)
	r.HandleFunc("/rd", func(w mux.ResponseWriter, r *mux.Message) {
		if r.Code == codes.POST {
			first, second, err := r.Options.Find(message.URIQuery)
			if err != nil {
				w.SetResponse(
					codes.BadRequest,
					message.TextPlain,
					bytes.NewReader([]byte("request should have uri-query option for ep!")),
				)
				return
			}

			for _, opt := range r.Options[first:second] {
				if strings.HasPrefix(string(opt.Value), "ep=") {

					remoteAddr := w.Client().RemoteAddr().String()
					_, err := net.ResolveIPAddr("ip", w.Client().RemoteAddr().String())

					if err != nil {
						if strings.HasPrefix(w.Client().RemoteAddr().String(), "[::1]") {
							remoteAddr = strings.Replace(
								w.Client().RemoteAddr().String(),
								"[::1]",
								"127.0.0.1",
								-1,
							)
						}
						w.SetResponse(
							codes.BadRequest,
							message.TextPlain,
							bytes.NewReader([]byte("invalid ip")),
						)
					}

					ms.cplist[remoteAddr] = string(opt.Value)[3:]

					w.SetResponse(
						codes.Created,
						message.TextPlain,
						bytes.NewReader([]byte(serverProxyAddr)),
					)
					return
				}
			}

			w.SetResponse(
				codes.BadRequest,
				message.TextPlain,
				bytes.NewReader([]byte("request should have uri-query option for ep!")),
			)
		} else if r.Code == codes.GET {
			first, second, err := r.Options.Find(message.URIQuery)
			if err != nil {
				w.SetResponse(
					codes.BadRequest,
					message.TextPlain,
					bytes.NewReader([]byte("request should have uri-query option for ep!!")),
				)
			}
			for _, opt := range r.Options[first:second] {
				if strings.HasPrefix(string(opt.Value), "addr=") {
					ep, ok := ms.cplist[string(opt.Value)[5:]]
					if !ok {
						w.SetResponse(
							codes.BadRequest,
							message.TextPlain,
							bytes.NewReader([]byte("invalid addr")),
						)
						return
					}

					w.SetResponse(
						codes.Content,
						message.TextPlain,
						bytes.NewReader([]byte(ep)),
					)
					return
				}
			}

			w.SetResponse(
				codes.BadRequest,
				message.TextPlain,
				bytes.NewReader([]byte("request should have uri-query option for ep!!!")),
			)

		} else {
			ms.handle(w, r)
		}

	})
	log.Fatal(coap.ListenAndServe("udp", ms.bind, r))
}

func NewMockServer(bind string) MockServerI {
	return &_mockServer{bind, map[string]string{}}
}
