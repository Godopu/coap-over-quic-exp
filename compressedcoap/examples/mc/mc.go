package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"scenarios/mock"
	"sync"
	"time"

	"github.com/plgd-dev/go-coap/v2/message"
)

const cnt = 100

func main() {
	recv := flag.Bool("recv", false, "send request?")
	s := flag.Int("s", 100, "size of message")
	flag.Parse()
	serverAdr := flag.Arg(0)

	b := make([]byte, *s)
	for i := 0; i < *s; i++ {
		b[i] = 'a'
	}
	payload := bytes.NewBuffer(b)

	if len(serverAdr) == 0 {
		log.Fatalln("please enter cp address")
	}

	var wg sync.WaitGroup
	if *recv {
		wg.Add(1)

		go func() {
			time.Sleep(time.Second * 5)
			wg.Done()
		}()

		mc, err := mock.NewMockClient(0, 600)
		if err != nil {
			panic(err)
		}

		_, err = mc.Connect(serverAdr)
		if err != nil {
			panic(err)
		}

	} else {
		wg.Add(cnt * 1)
		for i := 0; i < 1; i++ {
			go sendRequest(serverAdr, &wg, payload)
		}
	}

	wg.Wait()
}

func sendRequest(serverAdr string, wg *sync.WaitGroup, payload *bytes.Buffer) {
	mc, err := mock.NewMockClient(0, 600)
	if err != nil {
		panic(err)
	}

	location, err := mc.Connect(serverAdr)
	if err != nil {
		panic(err)
	}

	for i := 0; i < cnt; i++ {
		go func() {
			resp, err := mc.Client().Put(
				context.Background(),
				location,
				message.TextPlain,
				bytes.NewReader(payload.Bytes()),
			)
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(body))

			time.Sleep(time.Millisecond * 100)
			wg.Done()
		}()
	}
}
