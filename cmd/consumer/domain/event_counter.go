package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Julia-Marcal/eventcounter/pkg/logger"
)

type EventCounter struct {
	mu        sync.Mutex
	counters  map[string]map[string]int
	processed map[string]bool
}

func NewEventCounter() *EventCounter {
	return &EventCounter{
		counters:  make(map[string]map[string]int),
		processed: make(map[string]bool),
	}
}

func (c *EventCounter) Created(ctx context.Context, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.counters["created"] == nil {
		c.counters["created"] = make(map[string]int)
	}
	c.counters["created"][userID]++

	logger.Success("(CREATE) Evento processado para usuário %s, total: %d", userID, c.counters["created"][userID])
	fmt.Println()
	return nil
}

func (c *EventCounter) Updated(ctx context.Context, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.counters["updated"] == nil {
		c.counters["updated"] = make(map[string]int)
	}
	c.counters["updated"][userID]++

	logger.Success("(UPDATE) Evento processado para usuário %s, total: %d", userID, c.counters["updated"][userID])
	fmt.Println()
	return nil
}

func (c *EventCounter) Deleted(ctx context.Context, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.counters["deleted"] == nil {
		c.counters["deleted"] = make(map[string]int)
	}
	c.counters["deleted"][userID]++

	logger.Success("(DELETE) Evento processado para usuário %s, total: %d", userID, c.counters["deleted"][userID])
	fmt.Println()
	return nil
}

func (c *EventCounter) IsProcessed(messageID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.processed[messageID]
}

func (c *EventCounter) MarkProcessed(messageID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processed[messageID] = true
}

type UserCount struct {
	UserID string `json:"user_id"`
	Count  int    `json:"count"`
}

func (c *EventCounter) SaveResults() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resultsDir := "results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório results: %w", err)
	}

	event_types := []string{"created", "updated", "deleted"}

	for _, event_type := range event_types {
		filename := filepath.Join(resultsDir, fmt.Sprintf("%s.json", event_type))
		data := c.counters[event_type]
		if data == nil {
			data = make(map[string]int)
		}

		var userCounts []UserCount
		for userID, count := range data {
			userCounts = append(userCounts, UserCount{
				UserID: userID,
				Count:  count,
			})
		}

		json_data, err := json.MarshalIndent(userCounts, "", "  ")
		if err != nil {
			return fmt.Errorf("falha ao usar marshal nos dados de %s: %w", event_type, err)
		}

		err = os.WriteFile(filename, json_data, 0644)
		if err != nil {
			return fmt.Errorf("falha ao escrever no arquivo %s: %w", event_type, err)
		}

		logger.Success("Salvo %s com %d usuários", filename, len(data))
	}

	return nil
}
