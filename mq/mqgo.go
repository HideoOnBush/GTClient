package mq

import (
	"encoding/json"
	"log"
	"time"
)

func TestChannel() {
	var (
		conf = Conf{
			User: "kwq",
			Pwd:  "123456",
			Addr: "127.0.0.1",
			Port: "5672",
		}

		exchangeName = "user.register.direct"
		queueName    = "user.register.queue"
		keyName      = "user.register.event"
	)

	if err := Init(conf); err != nil {
		log.Fatalf(" mq init err: %v", err)
	}

	ch := NewChannel()
	if err := ch.ExchangeDeclare(exchangeName, "direct"); err != nil {
		log.Fatalf("create exchange err: %v", err)
	}
	if err := ch.QueueDeclare(queueName); err != nil {
		log.Fatalf("create queue err: %v", err)
	}
	if err := ch.QueueBind(queueName, keyName, exchangeName); err != nil {
		log.Fatalf("bind queue err: %v", err)
	}
	//get data
	data, _ := json.Marshal(Line{
		Source:       "ss",
		SourceIsCore: false,
		SourceScene:  "test",
		Target:       "tt",
		TargetIsCore: false,
		TargetScene:  "test",
		Dependence:   "weak",
	})
	go func() {
		for {
			if err := ch.Publish(exchangeName, keyName, data); err != nil {
				log.Fatalf("publish msg err: %v", err)
			}
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Minute)
}
