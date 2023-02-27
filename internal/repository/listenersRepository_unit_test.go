package repository

import (
	"context"
	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListenersRepository_CreateListener(t *testing.T) {
	ctx := context.Background()
	notify := &model.Notification{
		Position: &model.Position{
			ID:   "testID",
			Name: "testName",
		},
	}

	err := testListenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, notify)
	require.Error(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.Error(t, err)
	err = testListenersRepository.RemoveListenerSL(notify)
	require.NoError(t, err)
	err = testListenersRepository.RemoveListenerTP(notify)
	require.NoError(t, err)
}

func TestListenersRepository_RemoveListener(t *testing.T) {
	ctx := context.Background()
	notify := &model.Notification{
		Position: &model.Position{
			ID:   "testID",
			Name: "testName",
		},
	}

	err := testListenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)

	err = testListenersRepository.RemoveListenerSL(notify)
	require.NoError(t, err)
	err = testListenersRepository.RemoveListenerTP(notify)
	require.NoError(t, err)

	err = testListenersRepository.RemoveListenerSL(notify)
	require.Error(t, err)
	err = testListenersRepository.RemoveListenerTP(notify)
	require.Error(t, err)

	err = testListenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)

	err = testListenersRepository.RemoveListenerSL(notify)
	require.NoError(t, err)
	err = testListenersRepository.RemoveListenerTP(notify)
	require.NoError(t, err)
}

func TestListenersRepository_SendPrices(t *testing.T) {
	ctx := context.Background()
	notify := &model.Notification{
		Position: &model.Position{
			ID:         "testID",
			Name:       "testName",
			StopLoss:   100,
			TakeProfit: 100,
		},
	}

	err := testListenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)

	prices := []*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 101,
		},
	}

	testListenersRepository.SendPrices(prices)
	_, _ = testListenersRepository.ClosePosition(ctx)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, notify)
	require.Error(t, err)

	prices = []*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 99,
		},
	}

	testListenersRepository.SendPrices(prices)
	_, _ = testListenersRepository.ClosePosition(ctx)
	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.Error(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)

	testListenersRepository.SendPrices([]*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 101,
		},
	})

	testListenersRepository.SendPrices([]*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 11,
		},
	})

	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.Error(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, notify)
	require.Error(t, err)

	_, _ = testListenersRepository.ClosePosition(ctx)
	_, _ = testListenersRepository.ClosePosition(ctx)

	err = testListenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)
	err = testListenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
}
