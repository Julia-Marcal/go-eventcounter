package main

import (
	"context"
	"flag"
	"log"

	eventcounter "github.com/reb-felipe/eventcounter/pkg"
)

var (
	count        bool
	publish      bool
	size         int
	outputDir    string
	amqpUrl      string
	amqpExchange string
	declareQueue bool
)

func init() {
	flag.BoolVar(&count, "count", false, "Cria resumo das mensagens geradas")
	flag.StringVar(&outputDir, "count-out", ".", "Caminho de saida do resumo")
	flag.BoolVar(&publish, "publish", false, "Publica as mensagens no rabbitmq")
	flag.IntVar(&size, "size", 20, "Quantidade de mensagens geradas")
	flag.StringVar(&amqpUrl, "amqp-url", "amqp://guest:guest@localhost:5672", "URL do RabbitMQ")
	flag.StringVar(&amqpExchange, "amqp-exchange", "user-events", "Exchange do RabbitMQ")
	flag.BoolVar(&declareQueue, "amqp-declare-queue", false, "Declare fila no RabbitMQ")
	flag.Parse()
}

func main() {
	msgs := make([]*eventcounter.Message, size)
	for i := range msgs {
		msgs[i] = NewMessage()
	}

	if count {
		Write(outputDir, msgs)
	}

	if publish {
		if declareQueue {
			if err := Declare(); err != nil {
				log.Printf("Não foi possível declarar fila ou exchange, erro: %s", err.Error())
			}
		}

		if err := Publish(context.Background(), msgs); err != nil {
			log.Printf("Não foi possível publicar mensagem, erro: %s", err.Error())
		}
	}
}
