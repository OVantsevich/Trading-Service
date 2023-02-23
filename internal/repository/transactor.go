package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxTxKey struct{}

type TxFunc func(context.Context) error

// injects pgx.Tx into context
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, pgxTxKey{}, tx)
}

// retrieves pgx.Tx from context
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(pgxTxKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// PgxTransactor represents pgx transactor behavior
type PgxTransactor interface {
	WithinTransaction(ctx context.Context, txFn TxFunc) error
	WithinTransactionWithOptions(ctx context.Context, txFn TxFunc, opts pgx.TxOptions) error
}

type pgxTransactor struct {
	pool *pgxpool.Pool
}

// NewPgxTransactor builds new PgxTransactor
func NewPgxTransactor(p *pgxpool.Pool) PgxTransactor {
	return &pgxTransactor{pool: p}
}

// WithinTransaction runs WithinTransactionWithOptions with default tx options
func (t *pgxTransactor) WithinTransaction(ctx context.Context, txFunc TxFunc) error {
	return WithinTransaction(ctx, t.pool, txFunc)
}

// WithinTransactionWithOptions runs logic within transaction passing context with pgx.Tx injected into it,
// so you can retrieve it via PgxWithinTransactionRunner function Runner
func (t *pgxTransactor) WithinTransactionWithOptions(ctx context.Context, txFunc TxFunc, opts pgx.TxOptions) (err error) {
	return WithinTransactionWithOptions(ctx, t.pool, txFunc, opts)
}

// PgxQueryRunner represents query runner behavior
type PgxQueryRunner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, ident pgx.Identifier, cls []string, src pgx.CopyFromSource) (int64, error)
}

// PgxWithinTransactionRunner represents query runner retriever for pgx
type PgxWithinTransactionRunner interface {
	PgxQueryRunner
	Runner(ctx context.Context) PgxQueryRunner
	Pool() *pgxpool.Pool
}

type pgxWithinTransactionRunner struct {
	pool *pgxpool.Pool
}

// NewPgxWithinTransactionRunner builds new PgxWithinTransactionRunner
func NewPgxWithinTransactionRunner(p *pgxpool.Pool) PgxWithinTransactionRunner {
	return &pgxWithinTransactionRunner{pool: p}
}

// Runner extracts query runner from context, if pgx.Tx is injected into context it is returned and pgxpool.Pool otherwise
func (r *pgxWithinTransactionRunner) Runner(ctx context.Context) PgxQueryRunner {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}
	return r.pool
}

// Pool extracts *pgxpool.Pool from transaction runner
func (r *pgxWithinTransactionRunner) Pool() *pgxpool.Pool {
	return r.pool
}

// Exec calls pgxpool.Pool.Exec or pgx.Tx.Exec depending on execution context
func (r *pgxWithinTransactionRunner) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return r.Runner(ctx).Exec(ctx, sql, arguments...)
}

// Query calls pgxpool.Pool.Query or pgx.Tx.Query depending on execution context
func (r *pgxWithinTransactionRunner) Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error) {
	return r.Runner(ctx).Query(ctx, sql, optionsAndArgs...)
}

// QueryRow calls pgxpool.Pool.QueryRow or pgx.Tx.QueryRow depending on execution context
func (r *pgxWithinTransactionRunner) QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row {
	return r.Runner(ctx).QueryRow(ctx, sql, optionsAndArgs...)
}

// Begin calls pgxpool.Pool.Begin or pgx.Tx.Begin depending on execution context
func (r *pgxWithinTransactionRunner) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.Runner(ctx).Begin(ctx)
}

// SendBatch calls pgxpool.Pool.SendBatch or pgx.Tx.SendBatch depending on execution context
func (r *pgxWithinTransactionRunner) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return r.Runner(ctx).SendBatch(ctx, b)
}

// CopyFrom calls pgxpool.Pool.CopyFrom or pgx.Tx.CopyFrom depending on execution context
func (r *pgxWithinTransactionRunner) CopyFrom(ctx context.Context, ident pgx.Identifier, cls []string, src pgx.CopyFromSource) (int64, error) {
	return r.Runner(ctx).CopyFrom(ctx, ident, cls, src)
}

// PgxTransactionInitiator represents transaction initiator
type PgxTransactionInitiator interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// WithinTransaction runs WithinTransactionWithOptions with default tx options
func WithinTransaction(ctx context.Context, txInit PgxTransactionInitiator, txFunc TxFunc) error {
	return WithinTransactionWithOptions(ctx, txInit, txFunc, pgx.TxOptions{})
}

// WithinTransactionWithOptions runs logic within transaction passing context with pgx.Tx injected into it
func WithinTransactionWithOptions(ctx context.Context, txInit PgxTransactionInitiator, txFunc TxFunc, opts pgx.TxOptions) error {
	tx, err := txInit.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		var txErr error
		if err != nil {
			txErr = tx.Rollback(ctx)
		} else {
			txErr = tx.Commit(ctx)
		}

		if txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			err = txErr
		}
	}()

	err = txFunc(injectTx(ctx, tx))
	return err
}
