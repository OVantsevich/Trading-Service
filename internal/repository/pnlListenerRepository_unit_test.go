package repository

import (
	"context"
	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func TestPNLListenersRepository_AddPositions(t *testing.T) {
	ctx := context.Background()
	position := &model.Position{
		ID:            uuid.NewString(),
		User:          uuid.NewString(),
		Name:          uuid.NewString(),
		Amount:        100,
		PurchasePrice: 10.0,
		ShortPosition: true,
	}
	price := &model.Price{
		Name:          position.Name,
		SellingPrice:  rand.Float64() * 10,
		PurchasePrice: 0,
	}

	err := testPNLListenersRepository.AddPositions(ctx, []*model.Position{position}, map[string]*model.Price{position.Name: price})
	require.NoError(t, err)
}

func TestPNLListenersRepository_Send_Close(t *testing.T) {
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
			PurchasePrice: rand.Float64() * 100,
			ShortPosition: true,
		}
		price[position[i].Name] = &model.Price{
			Name:          position[i].Name,
			SellingPrice:  position[i].PurchasePrice - position[i].PurchasePrice/2,
			PurchasePrice: 0,
		}
	}

	err := testPNLListenersRepository.AddPositions(ctx, position, price)
	require.NoError(t, err)

	for _, p := range position {
		price[p.Name].SellingPrice *= 3
	}

	priceSlice := make([]*model.Price, len(price))
	j := 0
	for _, p := range price {
		priceSlice[j] = p
		j++
	}
	testPNLListenersRepository.SendPricesPNL(priceSlice)

	for _ = range price {
		_, err := testPNLListenersRepository.ClosePosition(ctx)
		require.NoError(t, err)
	}
	cancelContext, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(time.Second * 2)
		cancel()
	}()
	_, err = testPNLListenersRepository.ClosePosition(cancelContext)
	require.Error(t, err)
}
