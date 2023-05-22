package compressedcoap

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	gonet "net"
	"strings"
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/net"
	"github.com/plgd-dev/go-coap/v2/udp"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

type _CompressedCoAPServerProxy struct {
	quicBindAddr *gonet.UDPAddr
	coapBindPort *gonet.UDPAddr
	serverAddr   string
	coapConn     *client.ClientConn
	managers     map[string]*cpManager
}

func NewServerProxy(quicBindAddr, coapBindAddr *gonet.UDPAddr, serverAddr string) (ProxyI, error) {
	co, err := udp.Dial(serverAddr, udp.WithDialer(&gonet.Dialer{
		LocalAddr: &gonet.UDPAddr{
			Port: coapBindAddr.Port,
		},
		Control: reusePort,
	}))

	if err != nil {
		return nil, err
	}

	return &_CompressedCoAPServerProxy{
		quicBindAddr,
		coapBindAddr,
		serverAddr,
		co,
		map[string]*cpManager{},
	}, nil
}

func (sp *_CompressedCoAPServerProxy) isValidatedCP(cpAdr gonet.Addr) bool {
	resp, err := sp.coapConn.Get(context.Background(), "/rd", message.Option{
		ID:    message.URIQuery,
		Value: []byte(fmt.Sprintf("addr=%s", cpAdr.String())),
	})

	if err != nil {
		return false
	}

	return resp.Code() == codes.Content
}

type cpManager struct {
	managers map[uint64]StreamManager
	mutex    sync.Mutex
}

func (cm *cpManager) addStreamManager(path string, sm StreamManager) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.managers[getHashValue([]byte(path))] = sm
}

func (cm *cpManager) getStreamManager(path string) StreamManager {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	sm, ok := cm.managers[getHashValue([]byte(path))]
	if !ok {
		return nil
	}

	return sm
}

func (cm *cpManager) rmStreamManager(path string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.managers, getHashValue([]byte(path)))
}

func (cm *cpManager) listen(coapConn *client.ClientConn, quicConn quic.Connection) error {
	for {
		stream, err := quicConn.AcceptStream(context.Background())
		if err != nil {
			return err
		}

		// check the device is valid
		// if not valid reqeust, stream should be canceled.
		if !checkObjectValidation() {
			stream.Close()
		}

		var sm StreamManager
		if sm, err = NewStreamManager(coapConn, stream); err != nil {
			return err
		}

		sm.InitialFunc(
			func(msg *message.Message) error {
				if msg.Code != codes.POST {
					log.Println("first message should be post")
					return errors.New("first message should be post")
				}

				body, err := ioutil.ReadAll(msg.Body)
				if err != nil {
					return err
				}

				cm.addStreamManager(string(body), sm)

				// set close function
				sm.CloseFunc(func() {
					log.Println("sm is closed on sp")
					cm.rmStreamManager(string(body))
				})
				return nil
			},
		)
		sm.DefaultHandleFunc(func(m *message.Message) {
			m.Context = context.Background()
			resp, err := coapConn.Client().Do(m)
			if err != nil {
				return
			}

			sm.Queue(resp)
		})

		err = sm.Run()
		if err != nil {
			return err
		}
	}

}

func (sp *_CompressedCoAPServerProxy) runQUICServer() error {
	listener, err := quic.ListenAddrEarly(
		fmt.Sprintf("%s:%d", sp.quicBindAddr.IP, sp.quicBindAddr.Port),
		generateTLSConfig(),
		nil,
	)

	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}

		if !sp.isValidatedCP(conn.RemoteAddr()) {
			conn.CloseWithError(quic.ApplicationErrorCode(100), "invalid ip/port")
		} else {
			remoteAddr := conn.RemoteAddr().String()

			if strings.HasPrefix(remoteAddr, "[::1]") {
				remoteAddr = strings.Replace(
					remoteAddr,
					"[::1]",
					"127.0.0.1",
					-1,
				)
			}

			manager := &cpManager{
				managers: map[uint64]StreamManager{},
			}
			sp.managers[remoteAddr] = manager
			go manager.listen(sp.coapConn, conn)
		}
	}
}

func (sp *_CompressedCoAPServerProxy) runCoAPServer() error {
	lc := gonet.ListenConfig{
		Control: reusePort,
	}

	lp, err := lc.ListenPacket(
		context.Background(),
		"udp",
		fmt.Sprintf(":%d", sp.coapBindPort.Port),
	)

	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}

	conn := lp.(*gonet.UDPConn)
	// err = ipv4.NewPacketConn(conn).SetControlMessage(ipv4.FlagDst|ipv4.FlagInterface, true)
	// if err != nil {
	// 	log.Fatalf("set control msg failed: %v", err)
	// }

	rt := mux.NewRouter()
	rt.DefaultHandleFunc(func(w mux.ResponseWriter, r *mux.Message) {
		sp.defaultCoAPHandler(w, r)
	})

	s := udp.NewServer(udp.WithMux(rt))
	return s.Serve(net.NewUDPConn("udp", conn))
}

func (sp *_CompressedCoAPServerProxy) Run() error {

	go sp.runQUICServer()
	return sp.runCoAPServer()
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
