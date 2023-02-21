// Package repository price service
package repository

import (
	"context"
	"fmt"

	"Trading-Service/internal/model"

	psProto "github.com/OVantsevich/Price-Service/proto"
)

// PriceService entity
type PriceService struct {
	ctx    context.Context
	client psProto.PriceServiceClient
	stream psProto.PriceService_GetPricesClient
}

// NewPriceServiceRepository price service repository constructor
func NewPriceServiceRepository(pspp psProto.PriceServiceClient, ctx context.Context) *PriceService {
	ps := &PriceService{client: pspp, ctx: ctx}
	ps.subscribe()
	return ps
}

func (ps *PriceService) subscribe() (err error) {
	ps.stream, err = ps.client.GetPrices(ps.ctx)
	if err != nil {
		return fmt.Errorf("priceService - Sebscribe - GetPrices: %e", err)
	}
	return
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
