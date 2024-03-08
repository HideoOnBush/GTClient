package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

var QJEtcd *EtcdClient

type EtcdClient struct {
	*clientv3.Client
}

type EtcdConfig struct {
	Host    []string
	TimeOut time.Duration
}

func InitialEtcd(conf EtcdConfig) *EtcdClient {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Host,
		DialTimeout: conf.TimeOut,
	})
	if err != nil {
		log.Fatalf("Initialilze Etcd Client failed ,err=%v", err)
	}
	QJEtcd = &EtcdClient{cli}
	return QJEtcd
}
