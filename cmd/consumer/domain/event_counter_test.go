package domain

import (
	"context"
	"testing"
	"time"
)

func TestEventCounter(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

	err := counter.Created(ctx, "user1")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	err = counter.Created(ctx, "user1")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	err = counter.Updated(ctx, "user1")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	err = counter.Deleted(ctx, "user1")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	counter.MarkProcessed("test-id-1")
	if !counter.IsProcessed("test-id-1") {
		t.Error("Era esperado que a mensagem fosse marcada como processada")
	}

	if counter.IsProcessed("non-existent") {
		t.Error("Era esperado que mensagem inexistente não fosse processada")
	}
}

func TestDispatcher(t *testing.T) {
	counter := NewEventCounter()
	dispatcher := NewDispatcher(counter)
	defer dispatcher.Close()

	user_id, event_type, err := dispatcher.ParseRoutingKey("user123.event.created")
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}
	if user_id != "user123" {
		t.Errorf("Esperado user_id como 'user123', obtido '%s'", user_id)
	}
	if event_type != "created" {
		t.Errorf("Esperado event_type como 'created', obtido '%s'", event_type)
	}

	user_id, event_type, err = dispatcher.ParseRoutingKey("invalid.format")
	if user_id != "" || event_type != "" {
		t.Error("Esperado strings vazias para chave de roteamento inválida")
	}
}

func TestDispatcherIntegration(t *testing.T) {
	counter := NewEventCounter()
	dispatcher := NewDispatcher(counter)
	defer dispatcher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dispatcher.StartWorkers(ctx)

	msg1 := EventMessage{
		UserID:    "user1",
		EventType: "created",
		MessageID: "msg1",
	}

	msg2 := EventMessage{
		UserID:    "user1",
		EventType: "updated",
		MessageID: "msg2",
	}

	dispatcher.Dispatch(ctx, msg1)
	dispatcher.Dispatch(ctx, msg2)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	dispatcher.WaitForCompletion()
}
