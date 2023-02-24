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
		ID:   "testID",
		Name: "testName",
	}

	err := listenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerSL(ctx, notify)
	require.Error(t, err)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.Error(t, err)
	err = listenersRepository.RemoveListenerSL(notify)
	require.NoError(t, err)
	err = listenersRepository.RemoveListenerTP(notify)
	require.NoError(t, err)
}

func TestListenersRepository_RemoveListener(t *testing.T) {
	ctx := context.Background()
	notify := &model.Notification{
		ID:   "testID",
		Name: "testName",
	}

	err := listenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)

	err = listenersRepository.RemoveListenerSL(notify)
	require.NoError(t, err)
	err = listenersRepository.RemoveListenerTP(notify)
	require.NoError(t, err)

	err = listenersRepository.RemoveListenerSL(notify)
	require.Error(t, err)
	err = listenersRepository.RemoveListenerTP(notify)
	require.Error(t, err)

	err = listenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)

	err = listenersRepository.RemoveListenerSL(notify)
	require.NoError(t, err)
	err = listenersRepository.RemoveListenerTP(notify)
	require.NoError(t, err)
}

func TestListenersRepository_SendPrices(t *testing.T) {
	ctx := context.Background()
	notify := &model.Notification{
		ID:    "testID",
		Name:  "testName",
		Price: 100,
	}

	err := listenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)

	prices := []*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 101,
		},
	}

	listenersRepository.SendPrices(prices)
	_, _ = listenersRepository.ClosePosition(ctx)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerSL(ctx, notify)
	require.Error(t, err)

	prices = []*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 99,
		},
	}

	listenersRepository.SendPrices(prices)
	_, _ = listenersRepository.ClosePosition(ctx)
	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.Error(t, err)
	err = listenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)

	listenersRepository.SendPrices([]*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 101,
		},
	})

	listenersRepository.SendPrices([]*model.Price{
		&model.Price{
			Name:         "testName",
			SellingPrice: 11,
		},
	})

	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.Error(t, err)
	err = listenersRepository.CreateListenerSL(ctx, notify)
	require.Error(t, err)

	_, _ = listenersRepository.ClosePosition(ctx)
	_, _ = listenersRepository.ClosePosition(ctx)

	err = listenersRepository.CreateListenerTP(ctx, notify)
	require.NoError(t, err)
	err = listenersRepository.CreateListenerSL(ctx, notify)
	require.NoError(t, err)
}
