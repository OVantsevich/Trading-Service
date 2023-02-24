// Package repository price service
package repository

import (
	"context"
	"fmt"

	psProto "github.com/OVantsevich/Payment-Service/proto"
)

// PaymentService entity
type PaymentService struct {
	client psProto.PaymentServiceClient
}

// NewPaymentServiceRepository payment service repository constructor
func NewPaymentServiceRepository(pspp psProto.PaymentServiceClient) *PaymentService {
	ps := &PaymentService{client: pspp}
	return ps
}

// GetAccountID get user account id
func (p *PaymentService) GetAccountID(ctx context.Context, userID string) (string, error) {
	response, err := p.client.GetAccount(ctx, &psProto.GetAccountRequest{UserID: userID})
	if err != nil {
		return "", fmt.Errorf("paymentService - GetAccountID - GetAccount: %w", err)
	}
	return response.Account.ID, nil
}

// IncreaseAmount increase amount
func (p *PaymentService) IncreaseAmount(ctx context.Context, accountID string, amount float64) error {
	_, err := p.client.IncreaseAmount(ctx, &psProto.AmountRequest{AccountID: accountID, Amount: amount})
	if err != nil {
		return fmt.Errorf("paymentService - IncreaseAmount - IncreaseAmount: %w", err)
	}
	return nil
}

// DecreaseAmount decrease amount
func (p *PaymentService) DecreaseAmount(ctx context.Context, accountID string, amount float64) error {
	_, err := p.client.DecreaseAmount(ctx, &psProto.AmountRequest{AccountID: accountID, Amount: amount})
	if err != nil {
		return fmt.Errorf("paymentService - DecreaseAmount - DecreaseAmount: %w", err)
	}
	return nil
}
