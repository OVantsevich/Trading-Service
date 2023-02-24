// Package handler trading
package handler

import (
	"context"
	"time"

	"Trading-Service/internal/model"
	pr "Trading-Service/proto"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TradingService trading service
//
//go:generate mockery --name=TradingService --case=underscore --output=./mocks
type TradingService interface {
	CreatePosition(ctx context.Context, position *model.Position) (*model.Position, error)
	GetPositionByID(ctx context.Context, positionID string) (*model.Position, error)
	GetUserPositions(ctx context.Context, userID string) ([]*model.Position, error)
	SetStopLoss(ctx context.Context, positionID string, stopLoss float64, updated time.Time) error
	SetTakeProfit(ctx context.Context, positionID string, takeProfit float64, updated time.Time) error
	ClosePosition(ctx context.Context, positionID string, closed, updated time.Time) error
}

// Trading handler
type Trading struct {
	pr.UnimplementedTradingServiceServer
	service TradingService
}

// NewPrice constructor
func NewPrice(s TradingService) *Trading {
	return &Trading{service: s}
}

// OpenPosition open new position
func (t *Trading) OpenPosition(ctx context.Context, request *pr.OpenPositionRequest) (*pr.OpenPositionResponse, error) {
	position, err := t.service.CreatePosition(ctx, &model.Position{
		User:   request.UserID,
		Name:   request.Name,
		Amount: request.Amount,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"UserID": position.User,
			"Name":   position.Name,
			"Amount": position.Amount,
		}).Errorf("trading - OpenPosition - CreatePosition: %v", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pr.OpenPositionResponse{Position: positionToGRPC(position)}, nil
}

// ClosePosition close position
func (t *Trading) ClosePosition(ctx context.Context, request *pr.ClosePositionRequest) (*pr.Response, error) {
	err := t.service.ClosePosition(ctx, request.PositionID, time.Now(), time.Now())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"PositionID": request.PositionID,
		}).Errorf("trading - ClosePosition - ClosePosition: %v", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pr.Response{}, nil
}

// GetPositionByID get position by id
func (t *Trading) GetPositionByID(ctx context.Context, request *pr.GetPositionByIDRequest) (*pr.GetPositionByIDResponse, error) {
	position, err := t.service.GetPositionByID(ctx, request.PositionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"UserID": request.PositionID,
		}).Errorf("trading - GetPositionByID - GetPositionByID: %v", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pr.GetPositionByIDResponse{Position: positionToGRPC(position)}, nil
}

// GetUserPositions get positions by user id
func (t *Trading) GetUserPositions(ctx context.Context, request *pr.GetUserPositionsRequest) (*pr.GetUserPositionsResponse, error) {
	positions, err := t.service.GetUserPositions(ctx, request.UserID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"UserID": request.UserID,
		}).Errorf("trading - GetUserPositions - GetUserPositions: %v", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}
	resPos := make([]*pr.Position, len(positions))
	for i, p := range positions {
		resPos[i] = positionToGRPC(p)
	}
	return &pr.GetUserPositionsResponse{Position: resPos}, nil
}

// StopLoss set stop loss
func (t *Trading) StopLoss(ctx context.Context, request *pr.StopLossRequest) (*pr.Response, error) {
	err := t.service.SetStopLoss(ctx, request.PositionID, request.Price, time.Now())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"PositionID": request.PositionID,
		}).Errorf("trading - StopLoss - SetStopLoss: %v", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pr.Response{}, nil
}

// TakeProfit set take profit
func (t *Trading) TakeProfit(ctx context.Context, request *pr.TakeProfitRequest) (*pr.Response, error) {
	err := t.service.SetTakeProfit(ctx, request.PositionID, request.Price, time.Now())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"PositionID": request.PositionID,
		}).Errorf("trading - TakeProfit - SetTakeProfit: %v", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pr.Response{}, nil
}

func positionToGRPC(pos *model.Position) *pr.Position {
	prPos := &pr.Position{}
	if pos.StopLoss != 0 {
		prPos.StopLoss = &pos.StopLoss
	}
	if pos.TakeProfit != 0 {
		prPos.TakeProfit = &pos.TakeProfit
	}
	if !pos.Closed.IsZero() {
		closedUnix := pos.Closed.Unix()
		prPos.Closed = &closedUnix
	}
	if !pos.Created.IsZero() {
		createUnix := pos.Created.Unix()
		prPos.Created = &createUnix
	}
	return &pr.Position{
		Id:     pos.ID,
		Name:   pos.Name,
		Amount: pos.Amount,
	}
}
