package main

import (
	"GTClient/mq"
	"GTClient/quicServer"
	"context"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	conf := mq.Conf{
		User: "kwq",
		Pwd:  "123456",
		Addr: "127.0.0.1",
		Port: "5672",
	}
	ctx := context.Background()
	rabbitCh := mq.Initial(conf, &wg)
	_ = quicServer.InitializeServer(ctx, rabbitCh, &wg)

	//wg.Add(1)
	//_ = quicServer.In(&wg)

	wg.Wait()
}
