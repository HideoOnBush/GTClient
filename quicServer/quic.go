package quicServer

import (
	"GTClient/etcd"
	"GTClient/mq"
	"GTClient/quicServer/model"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/quic-go/quic-go"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"sync"
	"time"
)

type QServer struct {
	*mq.MqChannel
	*etcd.EtcdClient
}

type QConfig struct {
	Addr string
	Port int
}

const addr = "localhost:4242"
const message = "foobar"
const serviceId = 1
const procedureId = 1
const serviceName = "GetRelation"

type loggingWriter struct {
	io.Writer
	*mq.MqChannel
}

func (w loggingWriter) Write(b []byte) (int, error) {
	newMessageList := &model.LineList{}
	err := proto.Unmarshal(b, newMessageList)
	if err != nil {
		log.Fatalf("unmarshal in Write failed ,err=%v", err)
	}
	//var lines []*mq.Line
	//_ = json.Unmarshal(b, &lines)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	//fmt.Printf("Server: Got '%v'\n", newMessageList.Messages)
	for _, line := range newMessageList.Messages {
		w.DataC <- &mq.Line{
			Source:       line.Source,
			SourceIsCore: line.SourceIsCore,
			SourceScene:  line.SourceScene,
			Target:       line.Target,
			TargetIsCore: line.TargetIsCore,
			TargetScene:  line.TargetScene,
			Dependence:   line.Dependence,
			VisitCount:   line.VisitCount,
		}
	}
	return w.Writer.Write(b)
}

func In(wg *sync.WaitGroup) error {
	defer wg.Done()
	q := &QServer{}
	listener, err := quic.ListenAddr(addr, q.generateTLSConfig(), nil)
	if err != nil {
		return err
	}
	//defer listener.Close()

	conn, err := listener.Accept(context.Background())
	if err != nil {
		return err
	} else {
		log.Printf("accept conn and remoteAddr=%v", conn.RemoteAddr())
	}
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("err = %v", err)
		}
		//defer stream.Close()
		//Echo through the loggingWriter
		_, err = io.Copy(loggingWriter{stream, nil}, stream)
		//buf := make([]byte, len(message))
		//_, err = io.ReadFull(stream, buf)
		//if err != nil {
		//	return err
		//}
		//fmt.Printf("Server: Got '%s'\n", buf)
	}
	return err
}

func InitializeServer(ctx context.Context, mqC *mq.MqChannel, wg *sync.WaitGroup, client *etcd.EtcdClient, config QConfig) *QServer {
	q := &QServer{}
	q.EtcdClient = client
	q.MqChannel = mqC
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: config.Port})
	if err != nil {
		log.Printf("InitializeServer err = %v", err)
		os.Exit(1)
	}
	tr := quic.Transport{
		Conn: udpConn,
	}
	ln, err := tr.Listen(q.generateTLSConfig(), nil)
	//ln, err := quic.ListenAddr("127.0.0.1:1234", q.generateTLSConfig(), nil)
	if err != nil {
		log.Printf("InitializeServer err = %v", err)
		os.Exit(1)
	}
	wg.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer func() {
			cancel()
			wg.Done()
		}()
		go q.PutServiceForTracer(ctx, serviceName, serviceId, procedureId)
		go q.PutServiceForNginx(ctx, serviceName, config.Addr, config.Port)
		for {
			conn, err := ln.Accept(ctx)
			if err != nil {
				log.Printf("accept conn err = %v", err)
				continue
			} else {
				log.Printf("accept conn and remoteAddr=%v", conn.RemoteAddr())
			}
			go func() {
				for {
					str, err := conn.AcceptStream(ctx)
					if err != nil {
						if err.Error() == "Application error 0x0 (remote)" {
							log.Printf("HandleConn conn closed")
						} else {
							log.Printf("HandleConn conn err = %s", err.Error())
						}
						break
					} else {
						log.Printf("have strId = %v", str.StreamID())
					}
					go func() {
						_, err = io.Copy(loggingWriter{str, mqC}, str)
					}()
				}
			}()
		}
	}()
	return q
}

func (q *QServer) PutServiceForNginx(ctx context.Context, ServiceName string, Addr string, Port int) {
	key := fmt.Sprintf("/host/%s/hosts/%s:%d", ServiceName, Addr, Port)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			leaseResp, err := q.EtcdClient.Grant(context.Background(), 5)
			if err != nil {
				log.Printf("%s stop to extend lease in Etcd", ServiceName)
				return
			}
			_, err = q.EtcdClient.Put(context.Background(), key, "exist", clientv3.WithLease(leaseResp.ID))
			if err != nil {
				log.Printf("Put failed in Etcd,err=%v", err)
				return
			}
		case <-ctx.Done():
			log.Printf("%s stop to extend lease in Etcd", ServiceName)
			return
		}
	}
}

func (q *QServer) PutServiceForTracer(ctx context.Context, ServiceName string, serviceID uint32, procedureId uint16) {
	//put nginx addr and port
	key := fmt.Sprintf("/service/%s/%d/%d", ServiceName, serviceID, procedureId)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			leaseResp, err := q.EtcdClient.Grant(context.Background(), 5)
			if err != nil {
				log.Printf("%s stop to extend lease in Etcd", ServiceName)
				return
			}
			_, err = q.EtcdClient.Put(context.Background(), key, "exist", clientv3.WithLease(leaseResp.ID))
			if err != nil {
				log.Printf("Put failed in Etcd,err=%v", err)
				return
			}
		case <-ctx.Done():
			log.Printf("%s stop to extend lease in Etcd", ServiceName)
			return
		}
	}
}

func (q *QServer) generateTLSConfig() *tls.Config {
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
		NextProtos:   []string{"quic"},
	}
}
