// Package repository listeners repository
package repository

import (
	"context"
	"fmt"
	"github.com/OVantsevich/Trading-Service/internal/model"
	"sync"
)

// ListenersRepository listeners repository
type ListenersRepository struct {
	mu              sync.RWMutex
	closedPositions chan *model.Position
	listenersTP     map[string]map[string]chan *model.Price
	listenersSL     map[string]map[string]chan *model.Price
}

// NewListenersRepository constructor
func NewListenersRepository() *ListenersRepository {
	cpChan := make(chan *model.Position)
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
func (l *ListenersRepository) CreateListenerTP(ctx context.Context, position *model.Position) error {
	l.mu.Lock()
	lis, ok := l.listenersTP[position.Name]
	if !ok {
		l.listenersTP[position.Name] = make(map[string]chan *model.Price)
	}
	_, ok = lis[position.ID]
	if ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
	}
	channel := make(chan *model.Price, 1)
	go listener(ctx, channel, l.closedPositions, &(*position), func(price *model.Price, p *model.Position) bool {
		return price.SellingPrice >= p.TakeProfit != p.ShortPosition
	})
	l.listenersTP[position.Name][position.ID] = channel
	l.mu.Unlock()
	return nil
}

// CreateListenerSL create stop loss listener
//
//nolint:dupl //just because
func (l *ListenersRepository) CreateListenerSL(ctx context.Context, position *model.Position) error {
	l.mu.Lock()
	lis, ok := l.listenersSL[position.Name]
	if !ok {
		l.listenersSL[position.Name] = make(map[string]chan *model.Price)
	}
	_, ok = lis[position.ID]
	if ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
	}
	channel := make(chan *model.Price, 1)
	go listener(ctx, channel, l.closedPositions, &(*position), func(price *model.Price, p *model.Position) bool {
		return price.SellingPrice <= p.StopLoss != p.ShortPosition
	})
	l.listenersSL[position.Name][position.ID] = channel
	l.mu.Unlock()
	return nil
}

// RemoveListenerTP remove take profit listener
func (l *ListenersRepository) RemoveListenerTP(position *model.Position) error {
	l.mu.Lock()
	lis := l.listenersTP[position.Name]
	channel, ok := lis[position.ID]
	if !ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - RemoveListenerTP: listener with this name and positionID does't exist")
	}
	close(channel)
	delete(l.listenersTP[position.Name], position.ID)
	l.mu.Unlock()
	return nil
}

// RemoveListenerSL remove stop loss listener
func (l *ListenersRepository) RemoveListenerSL(position *model.Position) error {
	l.mu.Lock()
	lis := l.listenersSL[position.Name]
	channel, ok := lis[position.ID]
	if !ok {
		l.mu.Unlock()
		return fmt.Errorf("listenersRepository - RemoveListenerSL: listener with this name and positionID does't exist")
	}
	close(channel)
	delete(l.listenersSL[position.Name], position.ID)
	l.mu.Unlock()
	return nil
}

// SendPrices sending prices for all listeners
func (l *ListenersRepository) SendPrices(prices []*model.Price) {
	l.mu.RLock()
	for _, p := range prices {
		for _, lis := range l.listenersSL[p.Name] {
			lis <- &(*p)
		}
		for _, lis := range l.listenersTP[p.Name] {
			lis <- &(*p)
		}
	}
	l.mu.RUnlock()
}

// ClosePosition sync await for closed position from listeners
func (l *ListenersRepository) ClosePosition(ctx context.Context) (*model.Position, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("listenersRepository - ClosePosition: context canceld")
	case position := <-l.closedPositions:
		return position, nil
	}
}

func listener(ctx context.Context, cin chan *model.Price, cout chan *model.Position, position *model.Position,
	comp func(*model.Price, *model.Position) bool,
) {
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
			if comp(price, position) {
				position.SellingPrice = price.SellingPrice
				cout <- position
				skip(ctx, cin)
				return
			}
		}
	}
}

func skip(ctx context.Context, cin chan *model.Price) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-cin:
			if !ok {
				return
			}
		}
	}
}
