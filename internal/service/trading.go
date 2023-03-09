// Package service trading
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/OVantsevich/Trading-Service/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// stopLoss stop loss
const stopLoss = "stop_loss"

// takeProfit take profit
const takeProfit = "take_profit"

// closed
const closed = "closed"

// created
const created = "created"

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
	ClosePosition(ctx context.Context, id string, closed int64, sellingPrice float64, updated time.Time) (*model.Position, error)

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
	CreateListenerTP(ctx context.Context, position *model.Position) error
	CreateListenerSL(ctx context.Context, position *model.Position) error
	RemoveListenerTP(position *model.Position) error
	RemoveListenerSL(position *model.Position) error

	SendPrices(prices []*model.Price)
	ClosePosition(ctx context.Context) (*model.Position, error)
}

// ListenerPNL pnl listener
//
//go:generate mockery --name=ListenersRepository --case=underscore --output=./mocks
type ListenerPNL interface {
	AddPositions(ctx context.Context, positions []*model.Position, prices map[string]*model.Price) error
	RemovePosition(position *model.Position) error
	SendPricesPNL(prices []*model.Price)
	ClosePosition(ctx context.Context) (*model.Position, error)
}

// Trading trading service
type Trading struct {
	positionsRepository PositionsRepository
	priceService        PriceService
	paymentService      PaymentService
	listenersRepository ListenersRepository
	listenerPNL         ListenerPNL

	transactor repository.PgxTransactor
}

// NewTrading constructor
func NewTrading(ctx context.Context, lr ListenersRepository, pnllr ListenerPNL, pr PositionsRepository, pp PriceService, ps PaymentService, trx repository.PgxTransactor) *Trading {
	prc := &Trading{positionsRepository: pr, priceService: pp, paymentService: ps, listenersRepository: lr, listenerPNL: pnllr, transactor: trx}
	prc.startListener(ctx, getPricesListener).startListener(ctx, getNotificationListener).startListener(ctx, closePositionListener).startListener(ctx, closePositionListenerPNL)
	return prc
}

// CreatePosition open new position
func (t *Trading) CreatePosition(ctx context.Context, position *model.Position) (*model.Position, error) {
	var pos *model.Position
	err := t.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		var trxErr error
		position.ID = uuid.New().String()
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

		var accountID string
		accountID, trxErr = t.paymentService.GetAccountID(ctx, position.User)
		if trxErr != nil {
			trxErr = fmt.Errorf("trading - CreatePosition - GetAccount: %w", trxErr)
			return trxErr
		}

		position.PurchasePrice = price.PurchasePrice
		sum := position.Amount * price.PurchasePrice
		trxErr = t.paymentService.DecreaseAmount(ctx, accountID, sum)
		if trxErr != nil {
			trxErr = fmt.Errorf("trading - CreatePosition - DecreaseAmount: %w", trxErr)
			return trxErr
		}
		return nil
	})

	return pos, err
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
func (t *Trading) ClosePosition(ctx context.Context, positionID string, closed int64, updated time.Time) error {
	err := t.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		pos, trxErr := t.positionsRepository.GetPositionByID(ctx, positionID)
		if trxErr != nil {
			return fmt.Errorf("trading - ClosePosition - GetPositionByID: %w", trxErr)
		}

		response, trxErr := t.priceService.GetCurrentPrices(ctx, []string{pos.Name})
		if trxErr != nil {
			return fmt.Errorf("trading - ClosePosition - GetCurrentPrices: %w", trxErr)
		}
		price := response[pos.Name]

		pos, trxErr = t.positionsRepository.ClosePosition(ctx, positionID, closed, price.SellingPrice, updated)
		if trxErr != nil {
			return fmt.Errorf("trading - ClosePosition - ClosePosition: %w", trxErr)
		}

		var sum float64
		if !pos.ShortPosition {
			var accountID string
			accountID, trxErr = t.paymentService.GetAccountID(ctx, pos.User)
			if trxErr != nil {
				trxErr = fmt.Errorf("trading - ClosePosition - GetAccountID: %w", trxErr)
				return trxErr
			}

			sum = pos.Amount * price.SellingPrice
			trxErr = t.paymentService.IncreaseAmount(ctx, accountID, sum)
			if trxErr != nil {
				trxErr = fmt.Errorf("trading - ClosePosition - IncreaseAmount: %w", trxErr)
				return trxErr
			}
		} else {
			if pos.PurchasePrice >= price.SellingPrice {
				sum = pos.Amount * (pos.PurchasePrice + pos.PurchasePrice - price.SellingPrice)

				var accountID string
				accountID, trxErr = t.paymentService.GetAccountID(ctx, pos.User)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - ClosePosition - GetAccountID: %w", trxErr)
					return trxErr
				}

				trxErr = t.paymentService.IncreaseAmount(ctx, accountID, sum)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - ClosePosition - IncreaseAmount: %w", trxErr)
					return trxErr
				}
			} else {
				sum = pos.Amount * pos.PurchasePrice

				var accountID string
				accountID, trxErr = t.paymentService.GetAccountID(ctx, pos.User)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - ClosePosition - GetAccountID: %w", trxErr)
					return trxErr
				}

				trxErr = t.paymentService.IncreaseAmount(ctx, accountID, sum)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - ClosePosition - IncreaseAmount: %w", trxErr)
					return trxErr
				}

				sum = pos.Amount * (price.SellingPrice - pos.PurchasePrice)

				trxErr = t.paymentService.DecreaseAmount(ctx, accountID, sum)
				if trxErr != nil {
					trxErr = fmt.Errorf("trading - ClosePosition - DecreaseAmount: %w", trxErr)
					return trxErr
				}
			}
		}
		return nil
	})
	return err
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
			t.listenerPNL.SendPricesPNL(prices)
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
				err = t.listenersRepository.CreateListenerTP(ctx, notify.Position)
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - CreateListenerTP: %w", err)
					continue
				}
				err = t.priceService.UpdateSubscription([]string{notify.Name})
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - UpdateSubscription: %w", err)
				}
			case stopLoss:
				err = t.listenersRepository.CreateListenerSL(ctx, notify.Position)
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - CreateListenerSL: %w", err)
					continue
				}
				err = t.priceService.UpdateSubscription([]string{notify.Name})
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - UpdateSubscription: %w", err)
				}
			case closed:
				if notify.TakeProfit > 0.0 {
					err = t.listenersRepository.RemoveListenerTP(notify.Position)
					if err != nil {
						errChan <- fmt.Errorf("trading - getNotificationListener - RemoveListenerTP: %w", err)
						continue
					}
				}
				if notify.StopLoss > 0.0 {
					err = t.listenersRepository.RemoveListenerSL(notify.Position)
					if err != nil {
						errChan <- fmt.Errorf("trading - getNotificationListener - RemoveListenerSL: %w", err)
						continue
					}
				}
				err = t.listenerPNL.RemovePosition(notify.Position)
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - RemovePosition: %w", err)
					continue
				}
			case created:
				prices, err := t.priceService.GetCurrentPrices(ctx, []string{notify.Name})
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - GetCurrentPrices: %w", err)
					continue
				}
				err = t.listenerPNL.AddPositions(ctx, []*model.Position{notify.Position}, map[string]*model.Price{notify.Name: prices[notify.Name]})
				if err != nil {
					errChan <- fmt.Errorf("trading - getNotificationListener - AddPositions: %w", err)
					continue
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
			err = t.ClosePosition(ctx, notify.ID, time.Now().Unix(), time.Now())
			if err != nil {
				errChan <- err
			}
		}
	}
}

func closePositionListenerPNL(ctx context.Context, t *Trading, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
		default:
			notify, err := t.listenerPNL.ClosePosition(ctx)
			if err != nil {
				errChan <- fmt.Errorf("trading - closePositionListenerPNL - ClosePosition: %w", err)
				continue
			}
			err = t.ClosePosition(ctx, notify.ID, time.Now().Unix(), time.Now())
			if err != nil {
				errChan <- err
			}
		}
	}
}
