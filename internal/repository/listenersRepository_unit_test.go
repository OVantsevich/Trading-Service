package repository

import (
	"context"
	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListenersRepository_CreateListener(t *testing.T) {
	ctx := context.Background()
	position := &model.Position{
		ID:   "testID",
		Name: "testName",
	}

	err := testListenersRepository.CreateListenerSL(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, position)
	require.Error(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.Error(t, err)
	err = testListenersRepository.RemoveListenerSL(position)
	require.NoError(t, err)
	err = testListenersRepository.RemoveListenerTP(position)
	require.NoError(t, err)
}

func TestListenersRepository_RemoveListener(t *testing.T) {
	ctx := context.Background()
	position := &model.Position{
		ID:   "testID",
		Name: "testName",
	}

	err := testListenersRepository.CreateListenerSL(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.NoError(t, err)

	err = testListenersRepository.RemoveListenerSL(position)
	require.NoError(t, err)
	err = testListenersRepository.RemoveListenerTP(position)
	require.NoError(t, err)

	err = testListenersRepository.RemoveListenerSL(position)
	require.Error(t, err)
	err = testListenersRepository.RemoveListenerTP(position)
	require.Error(t, err)

	err = testListenersRepository.CreateListenerSL(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.NoError(t, err)

	err = testListenersRepository.RemoveListenerSL(position)
	require.NoError(t, err)
	err = testListenersRepository.RemoveListenerTP(position)
	require.NoError(t, err)
}

func TestListenersRepository_SendPrices(t *testing.T) {
	ctx := context.Background()
	position := &model.Position{
		ID:         "testID",
		Name:       "testName",
		StopLoss:   100,
		TakeProfit: 100,
	}

	err := testListenersRepository.CreateListenerSL(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.NoError(t, err)

	prices := []*model.Price{
		{
			Name:         "testName",
			SellingPrice: 101,
		},
	}

	testListenersRepository.SendPrices(prices)
	pos, _ := testListenersRepository.ClosePosition(ctx)

	if pos.StopLoss > 0.0 {
		testListenersRepository.RemoveListenerSL(pos)
	}
	if pos.TakeProfit > 0.0 {
		testListenersRepository.RemoveListenerTP(pos)
	}

	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, position)
	require.NoError(t, err)

	prices = []*model.Price{
		{
			Name:         "testName",
			SellingPrice: 99,
		},
	}

	testListenersRepository.SendPrices(prices)
	pos, _ = testListenersRepository.ClosePosition(ctx)

	if pos.StopLoss > 0.0 {
		testListenersRepository.RemoveListenerSL(pos)
	}
	if pos.TakeProfit > 0.0 {
		testListenersRepository.RemoveListenerTP(pos)
	}

	err = testListenersRepository.CreateListenerTP(ctx, position)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, position)
	require.NoError(t, err)
}
