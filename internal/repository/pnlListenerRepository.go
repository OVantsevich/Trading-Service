package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/OVantsevich/Trading-Service/internal/model"
)

type newPosition struct {
	newPos       *model.Position
	currentPrice *model.Price
}

type userListener struct {
	removePosition chan *model.Position
	addPosition    chan *newPosition
	updatePrices   chan *model.Price
}

// PNLListenersRepository listeners repository
type PNLListenersRepository struct {
	mu              sync.RWMutex
	pnlDown         chan *model.Position
	listenersPrices map[string]map[string]chan *model.Price
	userListeners   map[string]*userListener
}

// NewPNLListenersRepository constructor
func NewPNLListenersRepository() *PNLListenersRepository {
	pnlDown := make(chan *model.Position, 1)
	listenersPrices := make(map[string]map[string]chan *model.Price)
	userListeners := make(map[string]*userListener)
	return &PNLListenersRepository{
		pnlDown:         pnlDown,
		listenersPrices: listenersPrices,
		userListeners:   userListeners,
	}
}

// CreateListener create pnl listener
func (l *PNLListenersRepository) createListener(ctx context.Context, positions []*model.Position, prices map[string]*model.Price) error {
	l.userListeners[positions[0].User] = &userListener{
		removePosition: make(chan *model.Position),
		addPosition:    make(chan *newPosition),
		updatePrices:   make(chan *model.Price),
	}

	var ok bool
	for _, n := range positions {
		if n.ShortPosition {
			_, ok = l.listenersPrices[n.Name]
			if !ok {
				l.listenersPrices[n.Name] = make(map[string]chan *model.Price)
			}
			l.listenersPrices[n.Name][n.User] = l.userListeners[n.User].updatePrices
		}
	}

	go pnlListener(ctx, l.userListeners[positions[0].User], l.pnlDown)

	for _, p := range positions {
		if p.ShortPosition {
			pos := *p
			price := *prices[pos.Name]
			l.userListeners[positions[0].User].addPosition <- &newPosition{
				newPos:       &pos,
				currentPrice: &price,
			}
		}
	}
	return nil
}

// AddPositions add position
func (l *PNLListenersRepository) AddPositions(ctx context.Context, positions []*model.Position, prices map[string]*model.Price) error {
	l.mu.Lock()
	_, ok := l.userListeners[positions[0].User]
	if !ok {
		err := l.createListener(ctx, positions, prices)
		if err != nil {
			return fmt.Errorf("PNLListenersRepository - AddPositions - createListener: %w", err)
		}
	} else {
		for _, p := range positions {
			if p.ShortPosition {
				_, ok = l.listenersPrices[p.Name]
				if !ok {
					l.listenersPrices[p.Name] = make(map[string]chan *model.Price)
				}
				l.listenersPrices[p.Name][p.User] = l.userListeners[p.User].updatePrices

				pos := *p
				price := *prices[pos.Name]
				l.userListeners[positions[0].User].addPosition <- &newPosition{
					newPos:       &pos,
					currentPrice: &price,
				}
			}
		}
	}
	l.mu.Unlock()
	return nil
}

// RemovePosition remove position
func (l *PNLListenersRepository) RemovePosition(position *model.Position) error {
	if position.ShortPosition {
		l.mu.Lock()
		_, ok := l.userListeners[position.User]
		if !ok {
			return fmt.Errorf("PNLListenersRepository - RemovePosition: listener for this user does not exist")
		} else {
			_, ok = l.listenersPrices[position.Name]
			if !ok {
				return fmt.Errorf("PNLListenersRepository - RemovePosition: no position for this name exists")
			}
			_, ok = l.listenersPrices[position.Name][position.User]
			if ok {
				delete(l.listenersPrices[position.Name], position.User)
			}
			l.userListeners[position.User].removePosition <- position
		}
		l.mu.Unlock()
	}
	return nil
}

// SendPricesPNL sending prices for all users listeners
func (l *PNLListenersRepository) SendPricesPNL(prices []*model.Price) {
	l.mu.RLock()
	for _, p := range prices {
		for _, lis := range l.listenersPrices[p.Name] {
			//go func(c chan *model.Price, inp *model.Price) {
			//	c <- inp
			//}(lis, &model.Price{
			//	Name:          p.Name,
			//	SellingPrice:  p.SellingPrice,
			//	PurchasePrice: p.PurchasePrice,
			//})
			lis <- &(*p)
		}
	}
	l.mu.RUnlock()
}

// ClosePosition sync await for closed position from listeners
func (l *PNLListenersRepository) ClosePosition(ctx context.Context) (*model.Position, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("PNLListenersRepository - ClosePosition: context canceld")
	case position := <-l.pnlDown:
		return position, nil
	}
}

func pnlListener(ctx context.Context, userLis *userListener, cout chan *model.Position) {
	var sum float64

	var ok bool
	var prices = make(map[string]*model.Price)
	var positions = make(map[string]*model.Position)

	var updPrices *model.Price
	var removePosition *model.Position
	var addPosition *newPosition
	for {
		select {
		case <-ctx.Done():
			return
		case updPrices, ok = <-userLis.updatePrices:
			if !ok {
				return
			}
			price, ok := prices[updPrices.Name]
			if ok {
				price.SellingPrice = updPrices.SellingPrice
				price.PurchasePrice = updPrices.PurchasePrice
			} else {
				continue
			}
			for {
				sum = recalculate(positions, prices)
				if sum < 0.0 {
					for _, pos := range positions {
						cout <- pos
						delete(positions, pos.Name)
						delete(prices, pos.Name)
						break
					}
				} else {
					break
				}
			}
		case addPosition, ok = <-userLis.addPosition:
			if !ok {
				return
			}
			_, ok = positions[addPosition.newPos.Name]
			if !ok {
				positions[addPosition.newPos.Name] = addPosition.newPos
				prices[addPosition.newPos.Name] = addPosition.currentPrice
			}
			for {
				sum = recalculate(positions, prices)
				if sum < 0.0 {
					for _, pos := range positions {
						cout <- pos
						delete(positions, pos.Name)
						delete(prices, pos.Name)
						break
					}
				} else {
					break
				}
			}
		case removePosition, ok = <-userLis.removePosition:
			if !ok {
				return
			}
			_, ok = positions[removePosition.Name]
			if ok {
				delete(positions, removePosition.Name)
				delete(prices, removePosition.Name)
			}
		}
	}
}

func recalculate(positions map[string]*model.Position, prices map[string]*model.Price) float64 {
	var sum float64
	for _, pos := range positions {
		if pos.ShortPosition {
			sum += pos.Amount * (pos.PurchasePrice + pos.PurchasePrice - prices[pos.Name].SellingPrice)
		}
	}
	return sum
}
