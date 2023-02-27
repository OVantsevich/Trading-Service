package repository

import (
	"context"
	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestPNLListenersRepository_AddPositions(t *testing.T) {
	ctx := context.Background()
	position := &model.Position{
		ID:            uuid.NewString(),
		User:          uuid.NewString(),
		Name:          uuid.NewString(),
		Amount:        100,
		ShortPosition: 10,
	}
	price := &model.Price{
		Name:          position.Name,
		SellingPrice:  rand.Float64() * 10,
		PurchasePrice: 0,
	}

	err := testPNLListenersRepository.AddPositions(ctx, []*model.Position{position}, map[string]*model.Price{position.Name: price})
	require.NoError(t, err)
}

func TestPNLListenersRepository_Send_Remove_Close(t *testing.T) {
	ctx := context.Background()
	position := make([]*model.Position, 10)
	price := make(map[string]*model.Price)

	user := uuid.NewString()

	for i := range position {
		position[i] = &model.Position{
			ID:            uuid.NewString(),
			User:          user,
			Name:          uuid.NewString(),
			Amount:        100,
			ShortPosition: rand.Float64() * 100,
		}
		price[position[i].Name] = &model.Price{
			Name:          position[i].Name,
			SellingPrice:  position[i].ShortPosition - position[i].ShortPosition/2,
			PurchasePrice: 0,
		}
	}

	err := testPNLListenersRepository.AddPositions(ctx, position, price)
	require.NoError(t, err)

	for _, p := range position {
		price[p.Name].SellingPrice *= 3
	}

	testPNLListenersRepository.SendPricesPNL(price)

	for _ = range price {
		_, err := testPNLListenersRepository.ClosePosition(ctx)
		require.NoError(t, err)
	}
}
