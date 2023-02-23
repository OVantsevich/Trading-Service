// Package repository position
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Trading-Service/internal/model"

	"github.com/jackc/pgx/v5"
)

// Position postgres entity
type Position struct {
	PgxWithinTransactionRunner
	listenConn *pgx.Conn
}

// NewPositionRepository creating new Position repository
func NewPositionRepository(p PgxWithinTransactionRunner, ctx context.Context) (*Position, error) {
	repos := &Position{PgxWithinTransactionRunner: p}
	conn, err := repos.Pool().Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("position - startListen - Acquire: %w", err)
	}
	_, err = conn.Exec(ctx, "listen thresholds")
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("position - startListen - Exec: %w", err)
	}
	repos.listenConn = conn.Conn()

	return repos, nil
}

// CreatePosition create position
func (p *Position) CreatePosition(ctx context.Context, position *model.Position) (*model.Position, error) {
	row := p.QueryRow(ctx,
		`insert into positions (id, user, name, amount, created, updated) values ($1, $2, $3, $4, $5, $6) returning id;`,
		position.ID, position.User, position.Name, position.Amount, position.Created, position.Updated)
	err := row.Scan(&position.ID)
	if err != nil {
		return nil, fmt.Errorf("position - CreatePosition - Scan: %w", err)
	}

	return position, nil
}

// GetPositionByID get Position by id
func (p *Position) GetPositionByID(ctx context.Context, id string) (*model.Position, error) {
	pos := &model.Position{}
	row := p.QueryRow(ctx, `select id, user, name, amount, stop_loss, take_profit, closed, created
									from positions where id = $1`, id)
	err := row.Scan(
		&pos.ID, &pos.User, &pos.Name, &pos.Amount, &pos.StopLoss, &pos.TakeProfit, &pos.Created, &pos.Created)
	if err != nil {
		return nil, fmt.Errorf("position - GetPositionByLogin - Scan: %w", err)
	}

	return pos, nil
}

// UpdatePosition update position excluding thresholds
func (p *Position) UpdatePosition(ctx context.Context, position *model.Position) error {
	_, err := p.Exec(ctx, `update positions set amount=$1, updated=$2 where id=$3 and closed is not null;`,
		position.Amount, position.Updated, position.ID)
	if err != nil {
		return fmt.Errorf("position - UpdatePosition - Exec: %w", err)
	}

	return nil
}

// SetStopLoss set stop loss
func (p *Position) SetStopLoss(ctx context.Context, id string, stopLoss float64, updated time.Time) error {
	_, err := p.Exec(ctx, `update positions set stop_loss=$1, updated=$2 where id=$3 and deleted is not null;`,
		stopLoss, updated, id)
	if err != nil {
		return fmt.Errorf("position - SetStopLoss - Exec: %w", err)
	}

	return nil
}

// SetTakeProfit set take profit
func (p *Position) SetTakeProfit(ctx context.Context, id string, takeProfit float64, updated time.Time) error {
	_, err := p.Exec(ctx, `update positions set take_profit=$1, updated=$2 where id=$3 and deleted is not null;`,
		takeProfit, updated, id)
	if err != nil {
		return fmt.Errorf("position - SetTakeProfit - Exec: %w", err)
	}

	return nil
}

// ClosePosition close position
func (p *Position) ClosePosition(ctx context.Context, id string, closed, updated time.Time) (amount float64, err error) {
	row := p.QueryRow(ctx, "update positions set closed=$1, updated=$2 where id=$3 returning amount;",
		closed, updated, id)
	err = row.Scan(&amount)
	if err != nil {
		return 0, fmt.Errorf("position - ClosePosition - Exec: %w", err)
	}

	return amount, nil
}

// GetNotification get notification from listen/notify
func (p *Position) GetNotification(ctx context.Context) (*model.Notification, error) {
	msg, err := p.listenConn.WaitForNotification(ctx)
	if err != nil {
		return nil, fmt.Errorf("position - GetNotification - WaitForNotification: %w", err)
	}

	notify := &model.Notification{}
	err = json.Unmarshal([]byte(msg.Payload), &notify)
	if err != nil {
		return nil, fmt.Errorf("positions - GetNotification - Unmarshal: %w", err)
	}

	return notify, nil
}
