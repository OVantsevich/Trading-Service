package repository

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/OVantsevich/Trading-Service/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPosition_Create_Close_Position(t *testing.T) {
	var err error
	ctx := context.Background()
	var testData = []*model.Position{
		{
			ID:      uuid.NewString(),
			User:    uuid.NewString(),
			Name:    "name",
			Amount:  100,
			Created: time.Now(),
			Updated: time.Now(),
		},
	}

	for _, p := range testData {
		_, err = testPositionRepository.CreatePosition(ctx, p)
		require.NoError(t, err)
		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)

		_, err = testPositionRepository.CreatePosition(ctx, p)
		require.Error(t, err)

		wrongPos := *p
		wrongPos.ID = "wrongID"
		_, err = testPositionRepository.CreatePosition(ctx, &wrongPos)
		require.Error(t, err)

		var closed *model.Position
		closed, err = testPositionRepository.ClosePosition(ctx, p.ID, time.Now().Unix(), 0.0, time.Now())
		require.NoError(t, err)
		require.Equal(t, closed.Amount, p.Amount)

		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)

		_, err = testPositionRepository.ClosePosition(ctx, p.ID, time.Now().Unix(), 0.0, time.Now())
		require.Error(t, err)

		p.ID = uuid.New().String()
		_, err = testPositionRepository.CreatePosition(ctx, p)
		require.NoError(t, err)
		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)

		_, err = testPositionRepository.ClosePosition(ctx, p.ID, time.Now().Add(time.Second).Unix(), 0.0, time.Now())
		require.NoError(t, err)

		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)
	}
}

func TestPosition_GetPositionByID(t *testing.T) {
	var err error
	ctx := context.Background()
	var testData = []*model.Position{
		{
			ID:      uuid.NewString(),
			User:    uuid.NewString(),
			Name:    "name",
			Amount:  100,
			Created: time.Now(),
			Updated: time.Now(),
		},
	}

	for _, p := range testData {
		_, err = testPositionRepository.CreatePosition(ctx, p)
		require.NoError(t, err)
		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)

		getById, err := testPositionRepository.GetPositionByID(ctx, p.ID)
		require.NoError(t, err)
		require.Equal(t, getById.Name, p.Name)
		require.Equal(t, getById.User, p.User)
		require.Equal(t, getById.Amount, p.Amount)

		wrongPos := *p
		wrongPos.ID = "wrongID"
		_, err = testPositionRepository.GetPositionByID(ctx, wrongPos.ID)
		require.Error(t, err)

		_, err = testPositionRepository.ClosePosition(ctx, p.ID, time.Now().Add(time.Second).Unix(), 0.0, time.Now())
		require.NoError(t, err)

		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)
	}
}

func TestPosition_GetUserPositions(t *testing.T) {
	var err error
	ctx := context.Background()
	var testData = make([]*model.Position, 10)
	for i := range testData {
		testData[i] = &model.Position{
			ID:      uuid.NewString(),
			User:    "user",
			Name:    uuid.NewString(),
			Amount:  rand.Float64(),
			Created: time.Now(),
			Updated: time.Now(),
		}
	}

	for _, p := range testData {
		_, err = testPositionRepository.CreatePosition(ctx, p)
		require.NoError(t, err)
		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)
	}

	getById, err := testPositionRepository.GetUserPositions(ctx, testData[0].User)
	require.NoError(t, err)
	require.Equal(t, len(testData), len(getById))

	for _, p := range testData {
		_, err = testPositionRepository.ClosePosition(ctx, p.ID, time.Now().Unix(), 0.0, time.Now())
		require.NoError(t, err)

		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)
	}
}

func TestPosition_SL_TP_GetPosition(t *testing.T) {
	var err error
	ctx := context.Background()
	var testData = make([]*model.Position, 10)
	for i := range testData {
		testData[i] = &model.Position{
			ID:      uuid.NewString(),
			User:    "user",
			Name:    uuid.NewString(),
			Amount:  rand.Float64(),
			Created: time.Now(),
			Updated: time.Now(),
		}
	}

	for _, p := range testData {
		_, err = testPositionRepository.CreatePosition(ctx, p)
		require.NoError(t, err)
		_, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)
	}

	for _, p := range testData {
		cancelContext, cancel := context.WithCancel(ctx)
		go func() {
			time.Sleep(3)
			cancel()
		}()
		res, err := testPositionRepository.GetNotification(cancelContext)
		require.Error(t, err)
		require.Equal(t, (*model.Notification)(nil), res)

		_, err = testPositionRepository.ClosePosition(ctx, p.ID, time.Now().Unix(), 0.0, time.Now())
		require.NoError(t, err)
		res, err = testPositionRepository.GetNotification(ctx)
		require.NoError(t, err)
		require.Equal(t, p.ID, res.ID)
		cancelContext, cancel = context.WithCancel(ctx)
		go func() {
			time.Sleep(time.Second)
			cancel()
		}()
		res, err = testPositionRepository.GetNotification(cancelContext)
		require.Error(t, err)
	}

}
