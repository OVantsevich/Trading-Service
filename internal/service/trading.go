// Package service trading
package service

import (
	"Trading-Service/internal/repository"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"Trading-Service/internal/model"
)

// stopLoss stop loss
const stopLoss = "stop_loss"

// takeProfit take profit
const takeProfit = "take_profit"

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
	CreateListenerTP(ctx context.Context, notify *model.Notification) error
	CreateListenerSL(ctx context.Context, notify *model.Notification) error
	RemoveListenerTP(notify *model.Notification) error
	RemoveListenerSL(notify *model.Notification) error

	SendPrices(prices []*model.Price)
	ClosePosition(ctx context.Context) (*model.Notification, error)
}

// Trading trading service
type Trading struct {
	positionsRepository PositionsRepository
	listenersRepository ListenersRepository
	priceService        PriceService
	paymentService      PaymentService

	transactor repository.PgxTransactor
}

// NewPrices constructor
func NewPrices(ctx context.Context, lr ListenersRepository, pr PositionsRepository, pp PriceService, ps PaymentService, trx repository.PgxTransactor) *Trading {
	prc := &Trading{positionsRepository: pr, priceService: pp, paymentService: ps, listenersRepository: lr, transactor: trx}
	prc.startListener(ctx, getPricesListener).startListener(ctx, getNotificationListener).startListener(ctx, closePositionListener)
	return prc
}

func (t *Trading) startListener(ctx context.Context, listener func(*Trading, context.Context, chan error)) *Trading {
	errorChan := make(chan error)
	go func(t *Trading, ctx context.Context, errorChan chan error) {
		go listener(t, ctx, errorChan)
		for {
			select {
			case <-ctx.Done():
			case err := <-errorChan:
				logrus.Error(err)
			}
		}
	}(t, ctx, errorChan)
	return t
}

func getPricesListener(t *Trading, ctx context.Context, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
		default:
			prices, err := t.priceService.GetPrices()
			if err != nil {
				errChan <- fmt.Errorf("trading - getPricesListener - GetPrices: %w", err)
				continue
			}
			t.listenersRepository.SendPrices(prices)
		}
	}
}

func getNotificationListener(t *Trading, ctx context.Context, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
		default:
			notify, err := t.positionsRepository.GetNotification(ctx)
			if err != nil {
				errChan <- fmt.Errorf("trading - getNotificationListener - GetNotification: %w", err)
				continue
			}

			switch notify.Type {
			case takeProfit:
				if notify.Closed != nil {
					err = t.listenersRepository.RemoveListenerTP(notify)
					errChan <- fmt.Errorf("trading - getNotificationListener - RemoveListenerTP: %w", err)
				}
				err = t.listenersRepository.CreateListenerTP(ctx, notify)
				errChan <- fmt.Errorf("trading - getNotificationListener - CreateListenerTP: %w", err)
			case stopLoss:
				if notify.Closed != nil {
					err = t.listenersRepository.RemoveListenerSL(notify)
					errChan <- fmt.Errorf("trading - getNotificationListener - RemoveListenerSL: %w", err)
				}
				err = t.listenersRepository.CreateListenerSL(ctx, notify)
				errChan <- fmt.Errorf("trading - getNotificationListener - CreateListenerSL: %w", err)
			}
		}
	}
}

func closePositionListener(t *Trading, ctx context.Context, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
		default:
			notify, err := t.listenersRepository.ClosePosition(ctx)
			if err != nil {
				errChan <- fmt.Errorf("trading - closePositionListener - ClosePosition: %w", err)
				continue
			}

			err = t.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
				amount, trxErr := t.positionsRepository.ClosePosition(ctx, notify.ID, time.Now(), time.Now())
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - closePositionListener - ClosePosition: %w", trxErr)
					return trxErr
				}

				sum := amount * notify.Price
				var accountID string
				accountID, trxErr = t.paymentService.GetAccount(notify.User)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - closePositionListener - GetAccount: %w", trxErr)
					return trxErr
				}

				trxErr = t.paymentService.IncreaseAmount(accountID, sum)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - closePositionListener - IncreaseAmount: %w", trxErr)
					return trxErr
				}
				return nil
			})

		}
	}
}
