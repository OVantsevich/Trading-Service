// Package repository listeners repository
package repository

import (
	"context"
	"fmt"
	"sync"

	"Trading-Service/internal/model"
)

// ListenersRepository listeners repository
type ListenersRepository struct {
	mu              sync.RWMutex
	closedPositions chan *model.Notification
	listenersTP     map[string]map[string]chan float64
	listenersSL     map[string]map[string]chan float64
}

// NewListenersRepository constructor
func NewListenersRepository() *ListenersRepository {
	cpChan := make(chan *model.Notification)
	listenersTP := make(map[string]map[string]chan float64)
	listenersSL := make(map[string]map[string]chan float64)
	return &ListenersRepository{
		closedPositions: cpChan,
		listenersTP:     listenersTP,
		listenersSL:     listenersSL,
	}
}

// CreateListenerTP create take profit listener
//
//nolint:dupl //just because
func (l *ListenersRepository) CreateListenerTP(ctx context.Context, notify *model.Notification) error {
	l.mu.Lock()
	lis := l.listenersTP[notify.Name]
	_, ok := lis[notify.ID]
	if ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
	}
	channel := make(chan float64)
	go listenerTP(ctx, channel, l.closedPositions, notify)
	l.listenersTP[notify.Name][notify.ID] = channel
	l.mu.Unlock()
	return nil
}

// CreateListenerSL create stop loss listener
//
//nolint:dupl //just because
func (l *ListenersRepository) CreateListenerSL(ctx context.Context, notify *model.Notification) error {
	l.mu.Lock()
	lis := l.listenersSL[notify.Name]
	_, ok := lis[notify.ID]
	if ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
	}
	channel := make(chan float64)
	go listenerSL(ctx, channel, l.closedPositions, notify)
	l.listenersSL[notify.Name][notify.ID] = channel
	l.mu.Unlock()
	return nil
}

// RemoveListenerTP remove take profit listener
func (l *ListenersRepository) RemoveListenerTP(notify *model.Notification) error {
	l.mu.Lock()
	lis := l.listenersTP[notify.Name]
	channel, ok := lis[notify.ID]
	if !ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - RemoveListenerTP: listener with this name and positionID does't exist")
	}
	close(channel)
	delete(l.listenersTP[notify.Name], notify.ID)
	l.mu.Unlock()
	return nil
}

// RemoveListenerSL remove stop loss listener
func (l *ListenersRepository) RemoveListenerSL(notify *model.Notification) error {
	l.mu.Lock()
	lis := l.listenersSL[notify.Name]
	channel, ok := lis[notify.ID]
	if !ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - RemoveListenerSL: listener with this name and positionID does't exist")
	}
	close(channel)
	delete(l.listenersSL[notify.Name], notify.ID)
	l.mu.Unlock()
	return nil
}

// SendPrices sending prices for all listeners
func (l *ListenersRepository) SendPrices(prices []*model.Price) {
	l.mu.RLock()
	for _, p := range prices {
		for _, lis := range l.listenersSL[p.Name] {
			lis <- p.SellingPrice
		}
		for _, lis := range l.listenersTP[p.Name] {
			lis <- p.SellingPrice
		}
	}
	l.mu.RUnlock()
}

// ClosePosition sync await for closed position from listeners
func (l *ListenersRepository) ClosePosition(ctx context.Context) (*model.Notification, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("listenersRepository - ClosePosition: context canceld")
	case positionID := <-l.closedPositions:
		return positionID, nil
	}
}

func listenerTP(ctx context.Context, cin chan float64, cout chan *model.Notification, notify *model.Notification) {
	var price float64
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return
		case price, ok = <-cin:
			if !ok {
				return
			}
			if price <= notify.Price {
				notify.Price = price
				cout <- notify
				return
			}
		}
	}
}

func listenerSL(ctx context.Context, cin chan float64, cout chan *model.Notification, notify *model.Notification) {
	var price float64
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return
		case price, ok = <-cin:
			if !ok {
				return
			}
			if price >= notify.Price {
				notify.Price = price
				cout <- notify
				return
			}
		}
	}
}
