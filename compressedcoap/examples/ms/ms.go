package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"scenarios/mock"
	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/udp"
)

const cnt = 10

func main() {
	spAdr := flag.String("spAdr", "localhost:12126", "server proxy address")
	req := flag.Bool("r", false, "Do you want to send request??")
	b := flag.String("bind", ":5683", "bind address")
	flag.Parse()

	if *req {
		go func() {
			fmt.Println("Go")
			time.Sleep(time.Second * 5)
			sendRequest(*spAdr)
		}()
	}

	ms := mock.NewMockServer(*b)
	ms.Run()
}

func sendRequest(spAdr string) error {
	co, err := udp.Dial("localhost:5683")
	if err != nil {
		return err
	}

	resp, err := co.Client().Get(
		context.Background(),
		"/",
	)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	co2, err := udp.Dial(spAdr)
	if err != nil {
		return err
	}

	for i := 0; i < cnt; i++ {
		resp2, err := co2.Client().Put(
			context.Background(),
			"/75002/1",
			message.TextPlain,
			bytes.NewReader([]byte("control message")),
			message.Option{
				ID:    message.ProxyURI,
				Value: body,
			},
		)
		if err != nil {
			return err
		}

		body2, err := ioutil.ReadAll(resp2.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(body2))
	}

	return nil
}
