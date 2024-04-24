package mq

import (
	"encoding/json"
	"log"
	"sync"
)

type MqChannel struct {
	ExchangeName string
	QueueName    string
	KeyName      string
	MqC          *Channel
	DataC        chan *Line
}

func Initial(conf Conf, wg *sync.WaitGroup) *MqChannel {
	var m *MqChannel = &MqChannel{}
	m.ExchangeName = "user.register.direct"
	m.QueueName = "user.register.queue"
	m.KeyName = "user.register.event"

	if err := Init(conf); err != nil {
		log.Fatalf(" mq init err: %v", err)
	}

	ch := NewChannel()
	if err := ch.ExchangeDeclare(m.ExchangeName, "direct"); err != nil {
		log.Fatalf("create exchange err: %v", err)
	}
	if err := ch.QueueDeclare(m.QueueName); err != nil {
		log.Fatalf("create queue err: %v", err)
	}
	if err := ch.QueueBind(m.QueueName, m.KeyName, m.ExchangeName); err != nil {
		log.Fatalf("bind queue err: %v", err)
	}
	m.MqC = ch
	m.DataC = make(chan *Line, conf.ChanSize)
	wg.Add(1)
	go m.dataFromChannelToRabbit(wg)
	return m
	//get data
	//data, _ := json.Marshal(Line{
	//	Source:       "ss",
	//	SourceIsCore: false,
	//	SourceScene:  "test",
	//	Target:       "tt",
	//	TargetIsCore: false,
	//	TargetScene:  "test",
	//	Dependence:   "weak",
	//})
	//go func() {
	//	for {
	//
	//		if err := ch.Publish(exchangeName, keyName, data); err != nil {
	//			log.Fatalf("publish msg err: %v", err)
	//		}
	//		time.Sleep(time.Second)
	//	}
	//}()
	//time.Sleep(time.Minute)
}

func (m *MqChannel) dataFromChannelToRabbit(wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range m.DataC {
		data, _ := json.Marshal(line)
		if err := m.MqC.Publish(m.ExchangeName, m.KeyName, data); err != nil {
			log.Fatalf("publish msg err: %v", err)
		}
	}
}
