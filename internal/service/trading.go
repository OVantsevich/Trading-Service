// Package service trading
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/OVantsevich/Trading-Service/internal/repository"

	"github.com/sirupsen/logrus"
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
	GetPositionByID(ctx context.Context, positionID string) (*model.Position, error)
	GetUserPositions(ctx context.Context, userID string) ([]*model.Position, error)
	UpdatePosition(ctx context.Context, position *model.Position) error
	SetStopLoss(ctx context.Context, positionID string, stopLoss float64, updated time.Time) error
	SetTakeProfit(ctx context.Context, positionID string, takeProfit float64, updated time.Time) error
	ClosePosition(ctx context.Context, positionID string, closed, updated time.Time) (*model.Position, error)

	GetNotification(ctx context.Context) (*model.Notification, error)
}

// PriceService grpc price service
//
//go:generate mockery --name=PriceService --case=underscore --output=./mocks
type PriceService interface {
	GetPrices() ([]*model.Price, error)
	UpdateSubscription(names []string) error

	GetCurrentPrices(ctx context.Context, names []string) (map[string]*model.Price, error)
}

// PaymentService grpc payment service
//
//go:generate mockery --name=PaymentService --case=underscore --output=./mocks
type PaymentService interface {
	GetAccountID(ctx context.Context, userID string) (string, error)
	IncreaseAmount(ctx context.Context, accountID string, amount float64) error
	DecreaseAmount(ctx context.Context, accountID string, amount float64) error
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

// NewTrading constructor
func NewTrading(ctx context.Context, lr ListenersRepository, pr PositionsRepository, pp PriceService, ps PaymentService, trx repository.PgxTransactor) *Trading {
	prc := &Trading{positionsRepository: pr, priceService: pp, paymentService: ps, listenersRepository: lr, transactor: trx}
	prc.startListener(ctx, getPricesListener).startListener(ctx, getNotificationListener).startListener(ctx, closePositionListener)
	return prc
}

// CreatePosition open new position
func (t *Trading) CreatePosition(ctx context.Context, position *model.Position) (*model.Position, error) {
	var pos *model.Position
	err := t.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		var trxErr error
		pos, trxErr = t.positionsRepository.CreatePosition(ctx, position)
		if trxErr != nil {
			return fmt.Errorf("trading - CreatePosition - CreatePosition: %w", trxErr)
		}

		var response map[string]*model.Price
		response, trxErr = t.priceService.GetCurrentPrices(ctx, []string{position.Name})
		if trxErr != nil {
			return fmt.Errorf("trading - CreatePosition - GetCurrentPrices: %w", trxErr)
		}
		price := response[position.Name]

		sum := position.Amount * price.PurchasePrice
		var accountID string
		accountID, trxErr = t.paymentService.GetAccountID(ctx, position.User)
		if trxErr != nil {
			trxErr = fmt.Errorf("trading - CreatePosition - GetAccount: %w", trxErr)
			return trxErr
		}

		trxErr = t.paymentService.DecreaseAmount(ctx, accountID, sum)
		if trxErr != nil {
			trxErr = fmt.Errorf("trading - CreatePosition - DecreaseAmount: %w", trxErr)
			return trxErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pos, nil
}

// GetPositionByID get position by id
func (t *Trading) GetPositionByID(ctx context.Context, positionID string) (*model.Position, error) {
	pos, err := t.positionsRepository.GetPositionByID(ctx, positionID)
	if err != nil {
		return nil, fmt.Errorf("trading - GetPositionByID - GetPositionByID: %w", err)
	}
	return pos, nil
}

// GetUserPositions get positions by user id
func (t *Trading) GetUserPositions(ctx context.Context, userID string) ([]*model.Position, error) {
	pos, err := t.positionsRepository.GetUserPositions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("trading - GetUserPositions - GetUserPositions: %w", err)
	}
	return pos, nil
}

// SetStopLoss set stop loss
func (t *Trading) SetStopLoss(ctx context.Context, positionID string, stopLoss float64, updated time.Time) error {
	err := t.positionsRepository.SetStopLoss(ctx, positionID, stopLoss, updated)
	if err != nil {
		return fmt.Errorf("trading - SetStopLoss - SetStopLoss: %w", err)
	}
	return nil
}

// SetTakeProfit set take profit
func (t *Trading) SetTakeProfit(ctx context.Context, positionID string, takeProfit float64, updated time.Time) error {
	err := t.positionsRepository.SetTakeProfit(ctx, positionID, takeProfit, updated)
	if err != nil {
		return fmt.Errorf("trading - SetTakeProfit - SetTakeProfit: %w", err)
	}
	return nil
}

// ClosePosition close position
func (t *Trading) ClosePosition(ctx context.Context, positionID string, closed, updated time.Time) error {
	err := t.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		pos, trxErr := t.positionsRepository.ClosePosition(ctx, positionID, closed, updated)
		if trxErr != nil {
			return fmt.Errorf("trading - ClosePosition - ClosePosition: %w", trxErr)
		}

		var response map[string]*model.Price
		response, trxErr = t.priceService.GetCurrentPrices(ctx, []string{pos.Name})
		if trxErr != nil {
			return fmt.Errorf("trading - ClosePosition - GetCurrentPrices: %w", trxErr)
		}
		price := response[pos.Name]

		sum := pos.Amount * price.SellingPrice
		var accountID string
		accountID, trxErr = t.paymentService.GetAccountID(ctx, pos.User)
		if trxErr != nil {
			trxErr = fmt.Errorf("trading - ClosePosition - GetAccount: %w", trxErr)
			return trxErr
		}

		trxErr = t.paymentService.IncreaseAmount(ctx, accountID, sum)
		if trxErr != nil {
			trxErr = fmt.Errorf("trading - ClosePosition - DecreaseAmount: %w", trxErr)
			return trxErr
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (t *Trading) startListener(ctx context.Context, listener func(context.Context, *Trading, chan error)) *Trading {
	errorChan := make(chan error)
	go func(ctx context.Context, t *Trading, errorChan chan error) {
		go listener(ctx, t, errorChan)
		for {
			select {
			case <-ctx.Done():
			case err := <-errorChan:
				logrus.Error(err)
			}
		}
	}(ctx, t, errorChan)
	return t
}

func getPricesListener(ctx context.Context, t *Trading, errChan chan error) {
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

func getNotificationListener(ctx context.Context, t *Trading, errChan chan error) {
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
					if err != nil {
						errChan <- fmt.Errorf("trading - getNotificationListener - RemoveListenerTP: %w", err)
						continue
					}
				}
				err = t.listenersRepository.CreateListenerTP(ctx, notify)
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - CreateListenerTP: %w", err)
					continue
				}
				err = t.priceService.UpdateSubscription([]string{notify.Name})
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - UpdateSubscription: %w", err)
				}
			case stopLoss:
				if notify.Closed != nil {
					err = t.listenersRepository.RemoveListenerSL(notify)
					if err != nil {
						errChan <- fmt.Errorf("trading - getNotificationListener - RemoveListenerSL: %w", err)
						continue
					}
				}
				err = t.listenersRepository.CreateListenerSL(ctx, notify)
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - CreateListenerSL: %w", err)
					continue
				}
				err = t.priceService.UpdateSubscription([]string{notify.Name})
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - UpdateSubscription: %w", err)
				}
			}
		}
	}
}

func closePositionListener(ctx context.Context, t *Trading, errChan chan error) {
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
				pos, trxErr := t.positionsRepository.ClosePosition(ctx, notify.ID, time.Now(), time.Now())
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - closePositionListener - ClosePosition: %w", trxErr)
					return trxErr
				}

				sum := pos.Amount * notify.Price
				var accountID string
				accountID, trxErr = t.paymentService.GetAccountID(ctx, notify.User)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - closePositionListener - GetAccount: %w", trxErr)
					return trxErr
				}

				trxErr = t.paymentService.IncreaseAmount(ctx, accountID, sum)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - closePositionListener - IncreaseAmount: %w", trxErr)
					return trxErr
				}
				return nil
			})
			if err != nil {
				errChan <- err
			}
		}
	}
}
