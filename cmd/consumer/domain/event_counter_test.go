package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// TESTES DE FUNCIONALIDADE BÁSICA
// =============================================================================

func TestEventCounter_BasicOperations(t *testing.T) {
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

func TestDispatcher_BasicOperations(t *testing.T) {
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

	msg := EventMessage{MessageID: "1", EventType: "created", UserID: "user1"}
	ctx := context.Background()

	err = counter.Created(ctx, msg.UserID)
	if err != nil {
		t.Errorf("Erro no processamento: %v", err)
	}

	counter.MarkProcessed(msg.MessageID)

	if !counter.IsProcessed("1") {
		t.Error("Era esperado que o evento fosse processado")
	}
}

// =============================================================================
// TESTES DE CONTEXT E TIMEOUT
// =============================================================================

func TestEventCounter_WithTimeoutContext(t *testing.T) {
	counter := NewEventCounter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	err := counter.Created(ctx, "user1")
	if err != nil {
		t.Errorf("EventCounter não deveria retornar erro: %v", err)
	}
}

func TestEventCounter_WithCancelledContext(t *testing.T) {
	counter := NewEventCounter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := counter.Created(ctx, "user1")
	if err != nil {
		t.Errorf("EventCounter não deveria retornar erro: %v", err)
	}
}

func TestDispatcher_WithTimeoutContext(t *testing.T) {
	counter := NewEventCounter()
	dispatcher := NewDispatcher(counter)
	defer dispatcher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	msg := EventMessage{
		UserID:    "user1",
		EventType: "created",
		MessageID: "msg1",
	}

	err := counter.Created(ctx, msg.UserID)
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	counter.MarkProcessed(msg.MessageID)

	if !counter.IsProcessed(msg.MessageID) {
		t.Error("Mensagem deveria estar processada")
	}
}

func TestDispatcher_WithCancelledContext(t *testing.T) {
	counter := NewEventCounter()
	dispatcher := NewDispatcher(counter)
	defer dispatcher.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := EventMessage{
		UserID:    "user1",
		EventType: "created",
		MessageID: "msg1",
	}

	err := counter.Created(ctx, msg.UserID)
	if err != nil {
		t.Errorf("Erro inesperado: %v", err)
	}

	counter.MarkProcessed(msg.MessageID)

	if !counter.IsProcessed(msg.MessageID) {
		t.Error("Mensagem deveria estar processada")
	}
}

// =============================================================================
// TESTES DE CONCORRÊNCIA
// =============================================================================

func TestEventCounter_ConcurrentOperations(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

	const numGoroutines = 5
	const numOperations = 3

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				userID := fmt.Sprintf("user%d", id)
				err := counter.Created(ctx, userID)
				if err != nil {
					t.Errorf("Erro na criação concorrente: %v", err)
				}
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				userID := fmt.Sprintf("user%d", id)
				err := counter.Updated(ctx, userID)
				if err != nil {
					t.Errorf("Erro na atualização concorrente: %v", err)
				}
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				messageID := fmt.Sprintf("msg-%d-%d", id, j)
				counter.MarkProcessed(messageID)
				if !counter.IsProcessed(messageID) {
					t.Errorf("Mensagem %s deveria estar marcada como processada", messageID)
				}
			}
		}(i)
	}

	wg.Wait()

	counter.mu.Lock()
	defer counter.mu.Unlock()

	if len(counter.counters) == 0 {
		t.Error("Nenhum contador foi criado")
	}
}

func TestDispatcher_ConcurrentDispatching(t *testing.T) {
	counter := NewEventCounter()

	const numMessages = 3
	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user%d", id)
			messageID := fmt.Sprintf("msg-%d", id)

			err := counter.Created(ctx, userID)
			if err != nil {
				t.Errorf("Erro no processamento: %v", err)
			}

			counter.MarkProcessed(messageID)
		}(i)
	}

	wg.Wait()

	processedCount := 0
	for i := 0; i < numMessages; i++ {
		if counter.IsProcessed(fmt.Sprintf("msg-%d", i)) {
			processedCount++
		}
	}

	if processedCount != numMessages {
		t.Errorf("Esperado %d mensagens processadas, obtido %d", numMessages, processedCount)
	}
}

// =============================================================================
// TESTES DE PERSISTÊNCIA
// =============================================================================

func TestEventCounter_SaveResults(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

	testUsers := []string{"user1", "user2", "user3"}

	for i, user := range testUsers {
		for j := 0; j <= i; j++ {
			counter.Created(ctx, user)
			counter.Updated(ctx, user)
			counter.Deleted(ctx, user)
		}
	}

	err := counter.SaveResults()
	if err != nil {
		t.Fatalf("Erro ao salvar resultados: %v", err)
	}

	eventTypes := []string{"created", "updated", "deleted"}
	for _, eventType := range eventTypes {
		filename := fmt.Sprintf("results/%s.json", eventType)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Arquivo %s não foi criado", filename)
		}
	}

	for _, eventType := range eventTypes {
		filename := fmt.Sprintf("results/%s.json", eventType)
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("Erro ao ler arquivo %s: %v", filename, err)
			continue
		}

		var results []UserCount
		err = json.Unmarshal(data, &results)
		if err != nil {
			t.Errorf("Erro ao fazer unmarshal do arquivo %s: %v", filename, err)
			continue
		}

		resultMap := make(map[string]int)
		for _, userCount := range results {
			resultMap[userCount.UserID] = userCount.Count
		}

		for i, user := range testUsers {
			expectedCount := i + 1
			if resultMap[user] != expectedCount {
				t.Errorf("Para %s em %s: esperado %d, obtido %d",
					user, eventType, expectedCount, resultMap[user])
			}
		}
	}

	for _, eventType := range eventTypes {
		filename := fmt.Sprintf("results/%s.json", eventType)
		os.Remove(filename)
	}
}

func TestEventCounter_SaveEmptyResults(t *testing.T) {
	counter := NewEventCounter()

	err := counter.SaveResults()
	if err != nil {
		t.Fatalf("Erro ao salvar resultados vazios: %v", err)
	}

	eventTypes := []string{"created", "updated", "deleted"}
	for _, eventType := range eventTypes {
		filename := fmt.Sprintf("results/%s.json", eventType)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Arquivo %s não foi criado para dados vazios", filename)
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("Erro ao ler arquivo %s: %v", filename, err)
			continue
		}

		var results map[string]int
		err = json.Unmarshal(data, &results)
		if err != nil {
			t.Errorf("Erro ao fazer unmarshal do arquivo %s: %v", filename, err)
			continue
		}

		if len(results) != 0 {
			t.Errorf("Esperado mapa vazio para %s, obtido %v", eventType, results)
		}

		os.Remove(filename)
	}
}

// =============================================================================
// TESTES DE INTEGRAÇÃO
// =============================================================================

func TestDispatcher_FullIntegration(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

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

	err := counter.Created(ctx, msg1.UserID)
	if err != nil {
		t.Errorf("Erro no processamento created: %v", err)
	}
	counter.MarkProcessed(msg1.MessageID)

	err = counter.Updated(ctx, msg2.UserID)
	if err != nil {
		t.Errorf("Erro no processamento updated: %v", err)
	}
	counter.MarkProcessed(msg2.MessageID)

	if !counter.IsProcessed("msg1") {
		t.Error("Mensagem msg1 deveria ter sido processada")
	}
	if !counter.IsProcessed("msg2") {
		t.Error("Mensagem msg2 deveria ter sido processada")
	}
}

// =============================================================================
// TESTES AVANÇADOS DE CONCORRÊNCIA E PERSISTÊNCIA
// =============================================================================

func TestEventCounter_SharedUserConcurrency(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

	const numGoroutines = 5
	const operations = 3

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				counter.Created(ctx, "shared_user")
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				counter.Updated(ctx, "shared_user")
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				counter.Deleted(ctx, "shared_user")
			}
		}(i)
	}

	wg.Wait()

	counter.mu.Lock()
	defer counter.mu.Unlock()

	expectedCount := numGoroutines * operations

	if counter.counters["created"]["shared_user"] != expectedCount {
		t.Errorf("Esperado %d eventos created, obtido %d",
			expectedCount, counter.counters["created"]["shared_user"])
	}

	if counter.counters["updated"]["shared_user"] != expectedCount {
		t.Errorf("Esperado %d eventos updated, obtido %d",
			expectedCount, counter.counters["updated"]["shared_user"])
	}

	if counter.counters["deleted"]["shared_user"] != expectedCount {
		t.Errorf("Esperado %d eventos deleted, obtido %d",
			expectedCount, counter.counters["deleted"]["shared_user"])
	}
}

func TestEventCounter_ExtensiveRaceConditions(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

	const numGoroutines = 10
	const operations = 3

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				counter.Created(ctx, "shared_user")
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				counter.Updated(ctx, "shared_user")
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				counter.Deleted(ctx, "shared_user")
			}
		}(i)
	}

	wg.Wait()

	counter.mu.Lock()
	defer counter.mu.Unlock()

	expectedCount := numGoroutines * operations

	if counter.counters["created"]["shared_user"] != expectedCount {
		t.Errorf("Esperado %d eventos created, obtido %d",
			expectedCount, counter.counters["created"]["shared_user"])
	}

	if counter.counters["updated"]["shared_user"] != expectedCount {
		t.Errorf("Esperado %d eventos updated, obtido %d",
			expectedCount, counter.counters["updated"]["shared_user"])
	}

	if counter.counters["deleted"]["shared_user"] != expectedCount {
		t.Errorf("Esperado %d eventos deleted, obtido %d",
			expectedCount, counter.counters["deleted"]["shared_user"])
	}
}

func TestEventCounter_ConcurrentPersistence(t *testing.T) {
	counter := NewEventCounter()
	ctx := context.Background()

	const numSaves = 3

	counter.Created(ctx, "user1")
	counter.Updated(ctx, "user1")
	counter.Deleted(ctx, "user1")

	var wg sync.WaitGroup

	for i := 0; i < numSaves; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := counter.SaveResults()
			if err != nil {
				t.Errorf("Erro no salvamento concorrente: %v", err)
			}
		}()
	}

	wg.Wait()

	eventTypes := []string{"created", "updated", "deleted"}
	for _, eventType := range eventTypes {
		filename := fmt.Sprintf("results/%s.json", eventType)
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("Erro ao ler arquivo %s após salvamentos concorrentes: %v", filename, err)
			continue
		}

		var results []UserCount
		err = json.Unmarshal(data, &results)
		if err != nil {
			t.Errorf("Erro ao fazer unmarshal do arquivo %s após salvamentos concorrentes: %v", filename, err)
			continue
		}

		resultMap := make(map[string]int)
		for _, userCount := range results {
			resultMap[userCount.UserID] = userCount.Count
		}

		if resultMap["user1"] != 1 {
			t.Errorf("Para user1 em %s: esperado 1, obtido %d", eventType, resultMap["user1"])
		}

		os.Remove(filename)
	}
}
