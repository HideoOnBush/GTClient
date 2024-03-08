package main

import (
	"GTClient/etcd"
	"GTClient/mq"
	"GTClient/quicServer"
	"context"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	conf := mq.Conf{
		User: "kwq",
		Pwd:  "123456",
		Addr: "127.0.0.1",
		Port: "5672",
	}
	eConf := etcd.EtcdConfig{
		Host:    []string{"127.0.0.1:2379"},
		TimeOut: 5 * time.Second,
	}
	qConf := quicServer.QConfig{
		Addr: "127.0.0.1",
		Port: 2234,
	}
	ctx := context.Background()
	rabbitCh := mq.Initial(conf, &wg)
	etcdClient := etcd.InitialEtcd(eConf)
	_ = quicServer.InitializeServer(ctx, rabbitCh, &wg, etcdClient, qConf)

	//wg.Add(1)
	//_ = quicServer.In(&wg)

	wg.Wait()
}
