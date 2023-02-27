package repository

//import (
//	"context"
//	"fmt"
//	"sync"
//
//	"github.com/OVantsevich/Trading-Service/internal/model"
//)
//
//// PNLListenersRepository listeners repository
//type PNLListenersRepository struct {
//	mu        sync.RWMutex
//	pnlDown   chan *model.Position
//	listeners map[string]map[string]chan *model.Price
//}
//
//// NewPNLListenersRepository constructor
//func NewPNLListenersRepository() *PNLListenersRepository {
//	pnlDown := make(chan *model.Position)
//	listenersPrice := make(map[string]map[string]chan *model.Price)
//	return &PNLListenersRepository{
//		pnlDown:   pnlDown,
//		listeners: listenersPrice,
//	}
//}
//
//// CreateListener create pnl listener
//func (l *PNLListenersRepository) CreateListener(ctx context.Context, notify *model.Position) error {
//	l.mu.Lock()
//	lis, ok := l.listeners[notify.Name]
//	if !ok {
//		l.listeners[notify.Name] = make(map[string]chan *model.Price)
//	}
//	_, ok = lis[notify.ID]
//	if ok {
//		l.mu.Unlock()
//		return fmt.Errorf("listenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
//	}
//	channel := make(chan *model.Price, 1)
//	sendNotify := *notify
//	sendNotify.Type = takeProfit
//	go listener(ctx, channel, l.closedPositions, &sendNotify, func(price *model.Price, position *model.Position) bool {
//		return price.SellingPrice >= position.TakeProfit == (0.0 == position.ShortPosition)
//	})
//	l.listenersTP[notify.Name][notify.ID] = channel
//	l.mu.Unlock()
//	return nil
//}
//
//// CreateListenerSL create stop loss listener
////
////nolint:dupl //just because
//func (l *ListenersRepository) CreateListenerSL(ctx context.Context, notify *model.Position) error {
//	l.mu.Lock()
//	lis, ok := l.listenersSL[notify.Name]
//	if !ok {
//		l.listenersSL[notify.Name] = make(map[string]chan *model.Price)
//	}
//	_, ok = lis[notify.ID]
//	if ok {
//		l.mu.Unlock()
//		return fmt.Errorf("listenersRepository - CreateListenerSL: listener with this name and positionID alredy exist")
//	}
//	channel := make(chan *model.Price, 1)
//	sendNotify := *notify
//	sendNotify.Type = stopLoss
//	go listener(ctx, channel, l.closedPositions, &sendNotify, func(price *model.Price, position *model.Position) bool {
//		return price.SellingPrice <= position.StopLoss == (0.0 == position.ShortPosition)
//	})
//	l.listenersSL[notify.Name][notify.ID] = channel
//	l.mu.Unlock()
//	return nil
//}
//
//// RemoveListenerTP remove take profit listener
//func (l *ListenersRepository) RemoveListenerTP(notify *model.Position) error {
//	l.mu.Lock()
//	lis := l.listenersTP[notify.Name]
//	channel, ok := lis[notify.ID]
//	if !ok {
//		l.mu.Unlock()
//		return fmt.Errorf("listenersRepository - RemoveListenerTP: listener with this name and positionID does't exist")
//	}
//	close(channel)
//	delete(l.listenersTP[notify.Name], notify.ID)
//	l.mu.Unlock()
//	return nil
//}
//
//// RemoveListenerSL remove stop loss listener
//func (l *ListenersRepository) RemoveListenerSL(notify *model.Position) error {
//	l.mu.Lock()
//	lis := l.listenersSL[notify.Name]
//	channel, ok := lis[notify.ID]
//	if !ok {
//		l.mu.Unlock()
//		return fmt.Errorf("listenersRepository - RemoveListenerSL: listener with this name and positionID does't exist")
//	}
//	close(channel)
//	delete(l.listenersSL[notify.Name], notify.ID)
//	l.mu.Unlock()
//	return nil
//}
//
//// SendPrices sending prices for all listeners
//func (l *ListenersRepository) SendPrices(prices []*model.Price) {
//	l.mu.RLock()
//	for _, p := range prices {
//		for _, lis := range l.listenersSL[p.Name] {
//			lis <- p
//		}
//		for _, lis := range l.listenersTP[p.Name] {
//			lis <- p
//		}
//	}
//	l.mu.RUnlock()
//}
//
//// ClosePosition sync await for closed position from listeners
//func (l *ListenersRepository) ClosePosition(ctx context.Context) (*model.Position, error) {
//	select {
//	case <-ctx.Done():
//		return nil, fmt.Errorf("listenersRepository - ClosePosition: context canceld")
//	case notify := <-l.closedPositions:
//		switch notify.Type {
//		case takeProfit:
//			l.RemoveListenerTP(notify)
//		case stopLoss:
//			l.RemoveListenerSL(notify)
//		}
//		return notify, nil
//	}
//}
//
//func pnlListener(ctx context.Context, cinPrice chan []*model.Price, cinNotify chan *model.Position, cout chan *model.Position, notify *model.Position) {
//	var sum float64
//
//	var prices map[string]*model.Price
//	var positions []*model.Position
//	var ok bool
//	var price *model.Price
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case price, ok = <-cinPrice:
//			if !ok {
//				return
//			}
//
//		}
//	}
//}
//
//func recalculate(positions []*model.Position, prices map[string]*model.Price) float64 {
//	var sum float64
//
//	for _, pos := range positions {
//		sum += pos.Amount *
//	}
//
//}
