// Package repository listeners repository
package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/OVantsevich/Trading-Service/internal/model"
)

// stopLoss stop loss
const stopLoss = "stop_loss"

// takeProfit take profit
const takeProfit = "take_profit"

// ListenersRepository listeners repository
type ListenersRepository struct {
	mu              sync.RWMutex
	closedPositions chan *model.Notification
	listenersTP     map[string]map[string]chan *model.Price
	listenersSL     map[string]map[string]chan *model.Price
}

// NewListenersRepository constructor
func NewListenersRepository() *ListenersRepository {
	cpChan := make(chan *model.Notification)
	listenersTP := make(map[string]map[string]chan *model.Price)
	listenersSL := make(map[string]map[string]chan *model.Price)
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
	lis, ok := l.listenersTP[notify.Name]
	if !ok {
		l.listenersTP[notify.Name] = make(map[string]chan *model.Price)
	}
	_, ok = lis[notify.ID]
	if ok {
		l.mu.Unlock()
		return fmt.Errorf("testListenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
	}
	channel := make(chan *model.Price, 1)
	sendNotify := *notify
	sendNotify.Type = takeProfit
	go listener(ctx, channel, l.closedPositions, &sendNotify, func(price *model.Price, notification *model.Notification) bool {
		return price.SellingPrice >= notification.Price
	})
	l.listenersTP[notify.Name][notify.ID] = channel
	l.mu.Unlock()
	return nil
}

// CreateListenerSL create stop loss listener
//
//nolint:dupl //just because
func (l *ListenersRepository) CreateListenerSL(ctx context.Context, notify *model.Notification) error {
	l.mu.Lock()
	lis, ok := l.listenersSL[notify.Name]
	if !ok {
		l.listenersSL[notify.Name] = make(map[string]chan *model.Price)
	}
	_, ok = lis[notify.ID]
	if ok {
		l.mu.Unlock()
		return fmt.Errorf("testListenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
	}
	channel := make(chan *model.Price, 1)
	sendNotify := *notify
	sendNotify.Type = stopLoss
	go listener(ctx, channel, l.closedPositions, &sendNotify, func(price *model.Price, notification *model.Notification) bool {
		return price.SellingPrice <= notification.Price
	})
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
		return fmt.Errorf("testListenersRepository - RemoveListenerTP: listener with this name and positionID does't exist")
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
		return fmt.Errorf("testListenersRepository - RemoveListenerSL: listener with this name and positionID does't exist")
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
			lis <- p
		}
		for _, lis := range l.listenersTP[p.Name] {
			lis <- p
		}
	}
	l.mu.RUnlock()
}

// ClosePosition sync await for closed position from listeners
func (l *ListenersRepository) ClosePosition(ctx context.Context) (*model.Notification, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("testListenersRepository - ClosePosition: context canceld")
	case notify := <-l.closedPositions:
		switch notify.Type {
		case takeProfit:
			l.RemoveListenerTP(notify)
		case stopLoss:
			l.RemoveListenerSL(notify)
		}
		return notify, nil
	}
}

func listener(ctx context.Context, cin chan *model.Price, cout chan *model.Notification, notify *model.Notification, comp func(price *model.Price, notification *model.Notification) bool) {
	var price *model.Price
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return
		case price, ok = <-cin:
			if !ok {
				return
			}
			if comp(price, notify) {
				notify.Price = price.SellingPrice
				cout <- notify
				return
			}
		}
	}
}
