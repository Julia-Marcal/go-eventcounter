package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Julia-Marcal/eventcounter/cmd/consumer/config"
	rabbitmq "github.com/Julia-Marcal/eventcounter/cmd/consumer/connection"
	domain "github.com/Julia-Marcal/eventcounter/cmd/consumer/domain"
	"github.com/Julia-Marcal/eventcounter/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

func startConsumer(ctx context.Context, messages <-chan amqp.Delivery, dispatcher *domain.Dispatcher, counter *domain.EventCounter) {
	timeout_ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				logger.System("Canal de mensagens fechado")
				return
			}

			cancel()
			timeout_ctx, cancel = context.WithTimeout(ctx, 5*time.Second)

			logger.Info("Corpo da mensagem bruta: %s", string(msg.Body))
			logger.Info("Chave de roteamento: %s", msg.RoutingKey)

			var event domain.Event
			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				logger.Error("Falha ao deserializar mensagem: %v", err)
				msg.Nack(false, false)
				continue
			}

			if counter.IsProcessed(event.ID) {
				logger.Warning("Evento %s já processado, ignorando", event.ID)
				msg.Ack(false)
				continue
			}

			user_id, event_type, err := dispatcher.ParseRoutingKey(msg.RoutingKey)
			if err != nil || user_id == "" || event_type == "" {
				logger.Error("Falha ao analisar chave de roteamento: %s", msg.RoutingKey)
				msg.Nack(false, false)
				continue
			}

			logger.Process("Processando evento: ID=%s, UserID=%s, Type=%s", event.ID, user_id, event_type)
			counter.MarkProcessed(event.ID)

			event_msg := domain.EventMessage{
				UserID:    user_id,
				EventType: strings.ToLower(event_type),
				MessageID: event.ID,
			}
			dispatcher.Dispatch(ctx, event_msg)

			msg.Ack(false)

		case <-timeout_ctx.Done():
			if timeout_ctx.Err() == context.DeadlineExceeded {
				logger.System("Nenhuma mensagem recebida por 5 segundos, encerrando...")
			}
			return

		case <-ctx.Done():
			logger.System("Contexto cancelado, encerrando consumer...")
			return
		}
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Falha ao carregar configuração: %v", err)
	}

	messages, conn, ch, err := rabbitmq.ConsumeMessages(cfg.RabbitMQConnString, cfg.QueueName)
	if err != nil {
		logger.Fatal("Falha ao consumir mensagens:", err)
	}
	defer conn.Close()
	defer ch.Close()

	counter := domain.NewEventCounter()
	dispatcher := domain.NewDispatcher(counter)
	defer dispatcher.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatcher.StartWorkers(ctx)

	logger.System(" [*] Aguardando mensagens. Serviço será encerrado após 5s sem mensagens")

	startConsumer(ctx, messages, dispatcher, counter)

	logger.System("Aguardando processamento de todas as mensagens...")
	dispatcher.WaitForCompletion()

	logger.System("Salvando resultados...")
	if err := counter.SaveResults(); err != nil {
		logger.Error("Erro ao salvar resultados: %v", err)
	} else {
		logger.Success("Resultados salvos com sucesso!")
	}

	logger.System("Serviço parado")
}
