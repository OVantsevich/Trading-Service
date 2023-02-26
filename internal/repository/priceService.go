// Package repository price service
package repository

import (
	"context"
	"fmt"

	"github.com/OVantsevich/Trading-Service/internal/model"

	psProto "github.com/OVantsevich/Price-Service/proto"
)

//go:generate mockery --name=PSC --case=underscore --output=./mocks
type PSC = psProto.PriceServiceClient

// PriceService entity
type PriceService struct {
	ctx    context.Context
	client psProto.PriceServiceClient
	stream psProto.PriceService_GetPricesClient
}

// NewPriceServiceRepository price service repository constructor
func NewPriceServiceRepository(ctx context.Context, pspp psProto.PriceServiceClient) (*PriceService, error) {
	ps := &PriceService{client: pspp, ctx: ctx}
	err := ps.subscribe()
	if err != nil {
		return nil, fmt.Errorf("priceService - NewPriceServiceRepository - subscribe : %w", err)
	}
	return ps, nil
}

func (ps *PriceService) subscribe() (err error) {
	ps.stream, err = ps.client.GetPrices(ps.ctx)
	if err != nil {
		return fmt.Errorf("priceService - Sebscribe - GetPrices: %e", err)
	}
	return
}

// GetCurrentPrices get current prices by names
func (ps *PriceService) GetCurrentPrices(ctx context.Context, names []string) (map[string]*model.Price, error) {
	grpcPrices, err := ps.client.GetCurrentPrices(ctx, &psProto.GetCurrentPricesRequest{Names: names})
	if err != nil {
		return nil, fmt.Errorf("priceService - GetCurrentPrices - GetCurrentPrices: %w", err)
	}
	prices := mapFromGRPC(grpcPrices.Prices)
	return prices, nil
}

// GetPrices get prices from price service
func (ps *PriceService) GetPrices() ([]*model.Price, error) {
	response, err := ps.stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("priceService - GetPrices - Recv: %e", err)
	}
	return fromGRPC(response.Prices), nil
}

// UpdateSubscription subscribe for new prices
func (ps *PriceService) UpdateSubscription(names []string) error {
	err := ps.stream.Send(&psProto.GetPricesRequest{Names: names})
	if err != nil {
		return fmt.Errorf("priceService - UpdateSubscription - Send: %e", err)
	}
	return nil
}

func fromGRPC(recv []*psProto.Price) []*model.Price {
	result := make([]*model.Price, len(recv))
	for i, p := range recv {
		result[i] = &model.Price{
			Name:          p.Name,
			SellingPrice:  p.SellingPrice,
			PurchasePrice: p.PurchasePrice,
		}
	}
	return result
}

func mapFromGRPC(recv map[string]*psProto.Price) map[string]*model.Price {
	result := make(map[string]*model.Price, len(recv))
	for i, p := range recv {
		result[i] = &model.Price{
			Name:          p.Name,
			SellingPrice:  p.SellingPrice,
			PurchasePrice: p.PurchasePrice,
		}
	}
	return result
}
