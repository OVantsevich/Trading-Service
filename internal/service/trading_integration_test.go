package service

import (
	"context"
	"github.com/OVantsevich/Trading-Service/internal/model"
	"github.com/OVantsevich/Trading-Service/internal/repository"
	"github.com/OVantsevich/Trading-Service/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTrading_Listeners(t *testing.T) {
	priceService := mocks.NewPriceService(t)
	paymentService := mocks.NewPaymentService(t)

	user := uuid.NewString()
	position := make([]*model.Position, 10)
	price := make(map[string]*model.Price, len(position))
	priceSlice := make([]*model.Price, len(position))
	for i := range position {
		position[i] = &model.Position{
			ID:            uuid.NewString(),
			User:          user,
			Name:          uuid.NewString(),
			Amount:        100,
			ShortPosition: i%2 == 1,
			PurchasePrice: float64(i%2) * 30.0,
		}
		price[position[i].Name] = &model.Price{
			Name:          position[i].Name,
			SellingPrice:  20,
			PurchasePrice: 0,
		}
		priceSlice[i] = price[position[i].Name]
	}

	paymentService.On("IncreaseAmount", mock.AnythingOfType(""), mock.AnythingOfType("string"), mock.AnythingOfType("float64")).Maybe().Return(nil)
	paymentService.On("GetAccountID", mock.AnythingOfType(""), mock.AnythingOfType("string")).Maybe().Return("", nil)
	paymentService.On("DecreaseAmount", mock.AnythingOfType(""), mock.AnythingOfType("string"), mock.AnythingOfType("float64")).Maybe().Return(nil)

	until := make(chan time.Time)
	priceService.On("GetPrices").Maybe().WaitUntil(until).Return(priceSlice, nil)

	pnlListener := repository.NewPNLListenersRepository()
	listenerSLTP := repository.NewListenersRepository()

	ctx, cancel := context.WithCancel(context.Background())
	testTradingService = NewTrading(ctx, listenerSLTP, pnlListener, testPositionRepository, priceService, paymentService, repository.NewPgxTransactor(testPool))

	priceService.On("GetCurrentPrices", mock.AnythingOfType(""), mock.AnythingOfType("[]string")).Maybe().Return(
		price,
		nil,
	)
	var err error
	for _, p := range position {
		_, err = testTradingService.positionsRepository.CreatePosition(ctx, p)
		require.NoError(t, err)
	}

	priceService.On("UpdateSubscription", mock.AnythingOfType("[]string")).Return(nil).Maybe()

	for i := range position {
		if i%2 == 0 {
			err = testTradingService.positionsRepository.SetStopLoss(ctx, position[i].ID, 20, time.Now())
			require.NoError(t, err)
			err = testTradingService.positionsRepository.SetTakeProfit(ctx, position[i].ID, 80, time.Now())
			require.NoError(t, err)
		} else {
			err = testTradingService.positionsRepository.SetStopLoss(ctx, position[i].ID, 80, time.Now())
			require.NoError(t, err)
			err = testTradingService.positionsRepository.SetTakeProfit(ctx, position[i].ID, 10, time.Now())
			require.NoError(t, err)
			price[position[i].Name].SellingPrice = 5
			break
		}
	}
	logrus.Infof("%v", time.Now())
	until <- time.Now().Add(time.Millisecond)

	time.Sleep(time.Second * 1)
	poss, err := testTradingService.positionsRepository.GetUserPositions(ctx, position[0].User)
	require.NoError(t, err)
	var cls int
	for _, p := range poss {
		if p.Closed != 0 {
			cls++
		}
	}
	require.Equal(t, 2, cls)

	for i := range position {
		if i != 0 && i != 1 {
			if i%2 == 0 {
				err = testTradingService.positionsRepository.SetStopLoss(ctx, position[i].ID, 10, time.Now())
				require.NoError(t, err)
				err = testTradingService.positionsRepository.SetTakeProfit(ctx, position[i].ID, 80, time.Now())
				require.NoError(t, err)
			} else {
				err = testTradingService.positionsRepository.SetStopLoss(ctx, position[i].ID, 80, time.Now())
				require.NoError(t, err)
				err = testTradingService.positionsRepository.SetTakeProfit(ctx, position[i].ID, 0.01, time.Now())
				require.NoError(t, err)
			}
		}
		price[position[i].Name].SellingPrice = 0.1
	}
	until <- time.Now().Add(time.Millisecond)

	time.Sleep(time.Second)
	poss, err = testTradingService.positionsRepository.GetUserPositions(ctx, position[0].User)
	require.NoError(t, err)
	cls = 0
	for _, p := range poss {
		if p.Closed != 0 {
			cls++
		}
	}
	require.Equal(t, 6, cls)

	for i := range position {
		price[position[i].Name].SellingPrice = 60
	}
	until <- time.Now().Add(time.Millisecond)

	time.Sleep(time.Second)
	poss, err = testTradingService.positionsRepository.GetUserPositions(ctx, position[0].User)
	require.NoError(t, err)
	cls = 0
	for _, p := range poss {
		if p.Closed != 0 {
			cls++
		}
	}
	require.Equal(t, 10, cls)

	cancel()
}
