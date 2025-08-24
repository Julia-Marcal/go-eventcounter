package domain

import (
	"context"
	"strings"
	"sync"

	"github.com/Julia-Marcal/eventcounter/pkg/logger"
)

type EventMessage struct {
	UserID    string
	EventType string
	MessageID string
}

type Dispatcher struct {
	created_chan chan EventMessage
	updated_chan chan EventMessage
	deleted_chan chan EventMessage
	counter      *EventCounter
	wg           *sync.WaitGroup
}

func NewDispatcher(counter *EventCounter) *Dispatcher {
	return &Dispatcher{
		created_chan: make(chan EventMessage, 100),
		updated_chan: make(chan EventMessage, 100),
		deleted_chan: make(chan EventMessage, 100),
		counter:      counter,
		wg:           &sync.WaitGroup{},
	}
}

func (d *Dispatcher) ParseRoutingKey(routing_key string) (user_id, event_type string, err error) {
	parts := strings.Split(routing_key, ".")
	if len(parts) != 3 || parts[1] != "event" {
		return "", "", nil
	}

	user_id = parts[0]
	event_type = parts[2]
	return user_id, event_type, nil
}

func (d *Dispatcher) Dispatch(ctx context.Context, msg EventMessage) {
	d.wg.Add(1)

	switch msg.EventType {
	case "created":
		select {
		case d.created_chan <- msg:
			logger.Process("Evento (CREATED) enviado ao usuário %s", msg.UserID)
		case <-ctx.Done():
			d.wg.Done()
			return
		}
	case "updated":
		select {
		case d.updated_chan <- msg:
			logger.Process("Evento (UPDATED) enviado ao usuário %s", msg.UserID)
		case <-ctx.Done():
			d.wg.Done()
			return
		}
	case "deleted":
		select {
		case d.deleted_chan <- msg:
			logger.Process("Evento (DELETED) enviado ao usuário %s", msg.UserID)
		case <-ctx.Done():
			d.wg.Done()
			return
		}
	default:
		logger.Warning("Tipo de evento desconhecido: %s - usuário %s", msg.EventType, msg.UserID)
		d.wg.Done()
	}
}

func (d *Dispatcher) StartWorkers(ctx context.Context) {
	go func() {
		for {
			select {
			case msg := <-d.created_chan:
				if err := d.counter.Created(ctx, msg.UserID); err != nil {
					logger.Error("Erro ao processar evento (CREATED) para usuário %s: %v", msg.UserID, err)
				}
				d.wg.Done()
			case <-ctx.Done():
				logger.System("Worker (CREATED) parado")
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-d.updated_chan:
				if err := d.counter.Updated(ctx, msg.UserID); err != nil {
					logger.Error("Erro ao processar evento (UPDATED) para usuário %s: %v", msg.UserID, err)
				}
				d.wg.Done()
			case <-ctx.Done():
				logger.System("Worker (UPDATED) parado")
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-d.deleted_chan:
				if err := d.counter.Deleted(ctx, msg.UserID); err != nil {
					logger.Error("Erro ao processar evento (DELETED) para usuário %s: %v", msg.UserID, err)
				}
				d.wg.Done()
			case <-ctx.Done():
				logger.System("Worker (DELETED) parado")
				return
			}
		}
	}()
}

func (d *Dispatcher) WaitForCompletion() {
	d.wg.Wait()
}

func (d *Dispatcher) Close() {
	close(d.created_chan)
	close(d.updated_chan)
	close(d.deleted_chan)
}
