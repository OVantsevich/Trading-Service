// Package repository position
package repository

import (
	"context"
	"fmt"
	"time"

	"Trading-Service/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Position postgres entity
type Position struct {
	Pool *pgxpool.Pool
}

// NewPositionRepository creating new Position repository
func NewPositionRepository(pool *pgxpool.Pool) *Position {
	return &Position{Pool: pool}
}

// CreatePosition create position
func (r *Position) CreatePosition(ctx context.Context, position *model.Position) (*model.Position, error) {
	position.Created = time.Now()
	position.Updated = time.Now()
	_, err := r.Pool.Exec(ctx,
		`insert into positions (id, user, name, amount, stop_loss, take_profit, created, updated) values ($1, $2, $3, $4, $5, $6, $7. $8)
			 where ;`,
		position.ID, position.User, position.Name, position.Amount, position.StopLoss, position.TakeProfit, position.Created, position.Updated)
	if err != nil {
		return nil, fmt.Errorf("position - CreatePosition - Exec: %w", err)
	}

	return position, nil
}

// GetPositionByLogin get Position by login
func (r *Position) GetPositionByLogin(ctx context.Context, login string) (*model.Position, error) {
	pos := model.Position{}
	err := r.Pool.QueryRow(ctx, `select u.id, u.name, u.age, u.login, u.password, u.token, u.email, r.name
									from positions u
											 join roles r on r.id = u.role
									where u.login = $1 and u.deleted=false`, login).Scan(
		&pos.ID, &pos.Name, &pos.Age, &pos.Login, &pos.Password, &pos.Token, &pos.Email, &pos.Role)
	if err != nil {
		return nil, fmt.Errorf("Position - GetPositionByLogin - QueryRow: %w", err)
	}

	return &pos, nil
}

// GetPositionByID get Position by login
func (r *Position) GetPositionByID(ctx context.Context, id string) (*model.Position, error) {
	Position := model.Position{}
	err := r.Pool.QueryRow(ctx, `select u.id, u.name, u.age, u.login, u.password, u.token, u.email, r.name
									from Positions u
											 join roles r on r.id = u.role
									where u.id = $1 and u.deleted=false`, id).Scan(
		&Position.ID, &Position.Name, &Position.Age, &Position.Login, &Position.Password, &Position.Token, &Position.Email, &Position.Role)
	if err != nil {
		return nil, fmt.Errorf("Position - GetPositionByID - QueryRow: %w", err)
	}

	return &Position, nil
}

// UpdatePosition update position excluding thresholds
func (r *Position) UpdatePosition(ctx context.Context, position *model.Position) error {
	var idCheck int
	position.Updated = time.Now()
	err := r.Pool.QueryRow(ctx, "update positions set amount=$1, updated=$2 where id=$3 and deleted=false returning id",
		position.Amount, position.Updated, position.ID).Scan(&idCheck)
	if err != nil {
		return fmt.Errorf("position - UpdatePosition - Exec: %w", err)
	}

	return nil
}

// UpdateLowerThreshold update upper threshold
func (r *Position) UpdateLowerThreshold(ctx context.Context, id, token string) error {
	var idCheck int
	err := r.Pool.QueryRow(ctx, "update Positions set token=$1, updated=$2 where id=$3 and deleted=false returning id",
		token, time.Now(), id).Scan(&idCheck)
	if err != nil {
		return fmt.Errorf("Position - RefreshPosition - Exec: %w", err)
	}

	return nil
}

// UpdateUpperThreshold update lower threshold
func (r *Position) UpdateUpperThreshold(ctx context.Context, id, token string) error {
	var idCheck int
	err := r.Pool.QueryRow(ctx, "update Positions set token=$1, updated=$2 where id=$3 and deleted=false returning id",
		token, time.Now(), id).Scan(&idCheck)
	if err != nil {
		return fmt.Errorf("Position - RefreshPosition - Exec: %w", err)
	}

	return nil
}

// DeletePosition delete Position
func (r *Position) DeletePosition(ctx context.Context, id string) error {
	var idCheck int
	err := r.Pool.QueryRow(ctx, "update Positions set Deleted=true, updated=$1 where id=$2 and deleted=false returning id",
		time.Now(), id).Scan(&idCheck)
	if err != nil {
		return fmt.Errorf("Position - DeletePosition - Exec: %w", err)
	}

	return nil
}
