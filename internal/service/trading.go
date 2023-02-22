// Package service trading
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Trading-Service/internal/model"

	"github.com/google/uuid"
)

// batchSize number of messages read from mq
const batchSize = 100

// bufferSize number of messages stored for every grpc stream
const bufferSize = 1000

// PositionsRepository positions repository
//
//go:generate mockery --name=PositionsRepository --case=underscore --output=./mocks
type PositionsRepository interface {
	CreatePosition(ctx context.Context, position *model.Position) (*model.Position, error)
	GetPositionByID(ctx context.Context, id string) (*model.Position, error)
	UpdatePosition(ctx context.Context, position *model.Position) error
	SetStopLoss(ctx context.Context, id string, stopLoss float64, updated time.Time) error
	SetTakeProfit(ctx context.Context, id string, takeProfit float64, updated time.Time) error
	ClosePosition(ctx context.Context, id string, closed, updated time.Time) (float64, error)

	GetNotification(ctx context.Context) (*model.Notification, error)
}

// PriceService grpc price service
//
//go:generate mockery --name=PriceService --case=underscore --output=./mocks
type PriceService interface {
	GetPrices() ([]*model.Price, error)
	UpdateSubscription(names []string) error
}

// PaymentService grpc payment service
//
//go:generate mockery --name=PaymentService --case=underscore --output=./mocks
type PaymentService interface {
	GetAccount(userID string) (accountID string, err error)
	IncreaseAmount(accountID string, amount float64) error
	DecreaseAmount(accountID string, amount float64) error
}

// ListenersRepository pool of channels for TP and SL goroutines
//
//go:generate mockery --name=ListenersRepository --case=underscore --output=./mocks
type ListenersRepository interface {
	CreateListenerTP(positionID, name string, amount float64) error
	CreateListenerSL(positionID, name string, amount float64) error
	SendPrices(prices []*model.Price) error
	ClosePosition(ctx context.Context) (string, error)
}

// Trading trading service
type Trading struct {
	positionsRepository PositionsRepository
	listenersRepository ListenersRepository
	priceService        PriceService
	paymentService      PaymentService
}

// NewPrices constructor
func NewPrices(ctx context.Context, spr StreamPoolRepository, mq MQ, pp PriceService, startPosition string, end <-chan struct{}) *Prices {
	prc := &Prices{messageQueue: mq, priceProvider: pp, streamPool: spr}
	go prc.cycle(ctx, end, startPosition)
	return prc
}

// Subscribe allocating new channel for grpc stream with id and returning it
func (p *Prices) Subscribe(streamID uuid.UUID) chan *model.Price {
	streamChan := make(chan *model.Price, bufferSize)
	p.sMap.Store(streamID, streamChan)
	return streamChan
}

// UpdateSubscription clear previous stream subscription and creat new
func (p *Prices) UpdateSubscription(names []string, streamID uuid.UUID) error {
	streamChan, ok := p.sMap.Load(streamID)
	if !ok {
		return fmt.Errorf("not found")
	}
	p.streamPool.Delete(streamID)
	p.streamPool.Update(streamID, streamChan.(chan *model.Price), names)
	return nil
}

// DeleteSubscription delete grpc stream subscription and close it's chan
func (p *Prices) DeleteSubscription(streamID uuid.UUID) error {
	streamChan, ok := p.sMap.Load(streamID)
	if !ok {
		return fmt.Errorf("not found")
	}
	p.streamPool.Delete(streamID)
	close(streamChan.(chan *model.Price))
	p.sMap.Delete(streamID)
	return nil
}

func (p *Position) listen(ctx context.Context,
	endChan chan struct{}, errChan chan error, messages chan *model.Notification,
) {
	connection := p.listenConn.Conn()
	for {
		select {
		case <-endChan:
			p.listenConn.Release()
		default:
			msg, err := connection.WaitForNotification(ctx)
			if err != nil {
				errChan <- fmt.Errorf("position - listen - WaitForNotification: %w", err)
				return
			}

			notify := &model.Notification{}
			err = json.Unmarshal([]byte(msg.Payload), &notify)
			if err != nil {
				errChan <- fmt.Errorf("positions - listen - Unmarshal: %w", err)
				return
			}

			messages <- notify
		}
	}
}
