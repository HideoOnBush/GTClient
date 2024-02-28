package quicServer

import (
	"GTClient/mq"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"sync"
)

type QServer struct {
	*mq.MqChannel
}

const addr = "localhost:4242"

const message = "foobar"

type loggingWriter struct {
	io.Writer
	*mq.MqChannel
}

func (w loggingWriter) Write(b []byte) (int, error) {
	var lines []*mq.Line
	_ = json.Unmarshal(b, &lines)
	fmt.Printf("Server: Got '%v'\n", lines)
	for _, line := range lines {
		w.DataC <- line
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

func InitializeServer(ctx context.Context, mqC *mq.MqChannel, wg *sync.WaitGroup) *QServer {
	q := &QServer{}
	q.MqChannel = mqC
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 2234})
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
		defer wg.Done()
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
