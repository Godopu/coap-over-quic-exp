package main

import "fmt"

func main() {
	fmt.Println("compressed coap testbed")
}

// for simple coap test
// func main() {
// 	var mcWait sync.WaitGroup
// 	mcWait.Add(1)
// 	// mockDevice := mock.NewMockDevice(i, config.Params["mean"].(float64), reqPool, &wg)
// 	ms := mock.NewMockServer(":6262")
// 	go ms.Run()
// 	mc := mock.NewMockClient(0, 100, &mcWait, "localhost:6262")
// 	mc.Run()

// 	mcWait.Wait()
// }

// func main() {
// 	// init
// 	exptype := flag.String("exptype", "tcp", "experimentation type")
// 	Nd := flag.Int("Nd", 10, "Number of devices")
// 	Im := flag.Float64("Im", 0.5, "Interval of messages")
// 	bind := flag.String("bind", ":8080", "port to bind")
// 	spAdr := flag.String("spAdr", "localhost:8080", "server proxy address")
// 	proxytype := flag.String("proxytype", "cp", "proxy type (cp or sp)")
// 	expTime := flag.Int("expTime", 25, "experimentation time")
// 	flag.Parse()

// 	config.SetParam(*spAdr, *bind, *exptype, *Im, *Nd)

// 	var printFunc func(id int)
// 	switch *proxytype {
// 	case "cp":
// 		cpRun(*expTime)
// 		printFunc = timestamp.PrintStartStamp

// 	case "sp":
// 		spRun()
// 		printFunc = timestamp.PrintEndStamp
// 	}

// 	for i := 0; i < *Nd; i++ {
// 		printFunc(i)
// 	}
// }
