// Code generated by mockery v2.20.2. DO NOT EDIT.

package mocks

import (
	context "context"

	pgx "github.com/jackc/pgx/v5"
	mock "github.com/stretchr/testify/mock"

	repository "github.com/OVantsevich/Trading-Service/internal/repository"
)

// PgxTransactor is an autogenerated mock type for the PgxTransactor type
type PgxTransactor struct {
	mock.Mock
}

// WithinTransaction provides a mock function with given fields: ctx, txFn
func (_m *PgxTransactor) WithinTransaction(ctx context.Context, txFn repository.TxFunc) error {
	ret := _m.Called(ctx, txFn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, repository.TxFunc) error); ok {
		r0 = rf(ctx, txFn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WithinTransactionWithOptions provides a mock function with given fields: ctx, txFn, opts
func (_m *PgxTransactor) WithinTransactionWithOptions(ctx context.Context, txFn repository.TxFunc, opts pgx.TxOptions) error {
	ret := _m.Called(ctx, txFn, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, repository.TxFunc, pgx.TxOptions) error); ok {
		r0 = rf(ctx, txFn, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewPgxTransactor interface {
	mock.TestingT
	Cleanup(func())
}

// NewPgxTransactor creates a new instance of PgxTransactor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPgxTransactor(t mockConstructorTestingTNewPgxTransactor) *PgxTransactor {
	mock := &PgxTransactor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
