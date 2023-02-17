// Package main main
package main

import (
	"context"
	"fmt"
	"net"

	"Payment-Service/internal/config"
	"Payment-Service/internal/handler"
	"Payment-Service/internal/repository"
	"Payment-Service/internal/service"
	pr "Payment-Service/proto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.NewMainConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", cfg.Port))
	if err != nil {
		defer logrus.Fatalf("error while listening port: %e", err)
	}

	pool, err := dbConnection(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	trRepos := repository.NewTransaction(repository.NewPgxWithinTransactionRunner(pool))
	acRepos := repository.NewAccount(repository.NewPgxWithinTransactionRunner(pool))
	defer closePool(pool)

	trServ := service.NewTransaction(trRepos)
	acServ := service.NewAccount(acRepos)

	server := handler.NewUserHandler(trServ, acServ, repository.NewPgxTransactor(pool))

	ns := grpc.NewServer()
	pr.RegisterPaymentServiceServer(ns, server)

	if err = ns.Serve(listen); err != nil {
		defer logrus.Fatalf("error while listening server: %e", err)
	}
}

func dbConnection(cfg *config.MainConfig) (*pgxpool.Pool, error) {
	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword,
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)

	pool, err := pgxpool.New(context.Background(), pgURL)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration data: %v", err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database not responding: %v", err)
	}
	return pool, nil
}

func closePool(r interface{}) {
	p := r.(*pgxpool.Pool)
	if p != nil {
		p.Close()
	}
}
