// Package repository listeners repository
package repository

import (
	"Trading-Service/internal/model"
	"context"
	"fmt"
	"sync"
)

// ListenersRepository listeners repository
type ListenersRepository struct {
	mu              sync.RWMutex
	closedPositions chan string
	listenersTP     map[string]map[string]chan float64
	listenersSL     map[string]map[string]chan float64
}

// NewListenersRepository constructor
func NewListenersRepository() *ListenersRepository {
	cpChan := make(chan string)
	listenersTP := make(map[string]map[string]chan float64)
	listenersSL := make(map[string]map[string]chan float64)
	return &ListenersRepository{
		closedPositions: cpChan,
		listenersTP:     listenersTP,
		listenersSL:     listenersSL,
	}
}

func (l ListenersRepository) CreateListenerTP(positionID, name string, amount float64) error {
	l.mu.Lock()
	_, ok := l.listenersTP[name][positionID]
	if ok {
		return fmt.Errorf("listenersRepository - CreateListenerTP: listener with this name and positionID alredy exist")
	}

	l.mu.Unlock()
}

func (l ListenersRepository) CreateListenerSL(positionID, name string, amount float64) error {
	//TODO implement me
	panic("implement me")
}

func (l ListenersRepository) SendPrices(prices []*model.Price) error {
	//TODO implement me
	panic("implement me")
}

func (l ListenersRepository) ClosePosition(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

func listenerTP(ctx context.Context, cin chan float64, cout chan string, cost float64, positionID string) {
	var price float64
	var ok bool
	select {
	case <-ctx.Done():
		return
	case price, ok = <-cin:
		if price <= cost {
			if !ok {
				return
			}
			cout <- positionID
			return
		}
	default:
	}
}

func listenerSL(ctx context.Context, cin chan float64, cout chan string, cost float64, positionID string) {
	var price float64
	var ok bool
	select {
	case <-ctx.Done():
		return
	case price, ok = <-cin:
		if !ok {
			return
		}
		if price >= cost {
			cout <- positionID
			return
		}
	default:
	}
}
