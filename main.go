// Package main main
package main

import (
	"context"
	"fmt"
	"net"

	"github.com/OVantsevich/Trading-Service/internal/config"
	"github.com/OVantsevich/Trading-Service/internal/handler"
	"github.com/OVantsevich/Trading-Service/internal/repository"
	"github.com/OVantsevich/Trading-Service/internal/service"
	pr "github.com/OVantsevich/Trading-Service/proto"

	pasProto "github.com/OVantsevich/Payment-Service/proto"
	prsProto "github.com/OVantsevich/Price-Service/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
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
	positionRepository, err := repository.NewPositionRepository(ctx, repository.NewPgxWithinTransactionRunner(pool))
	if err != nil {
		logrus.Fatal(err)
	}
	defer closePool(pool)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	connPrice, err := grpc.Dial(fmt.Sprintf("%s:%s", cfg.PriceServiceHost, cfg.PriceServicePort), opts...)
	if err != nil {
		logrus.Fatal("Fatal Dial: ", err)
	}
	prsClient := prsProto.NewPriceServiceClient(connPrice)
	priceService, err := repository.NewPriceServiceRepository(ctx, prsClient)
	if err != nil {
		logrus.Fatal(err)
	}

	connPayment, err := grpc.Dial(fmt.Sprintf("%s:%s", cfg.PaymentServiceHost, cfg.PaymentServicePort), opts...)
	if err != nil {
		logrus.Fatal("Fatal Dial: ", err)
	}
	pasClient := pasProto.NewPaymentServiceClient(connPayment)
	paymentService := repository.NewPaymentServiceRepository(pasClient)

	listenerRepository := repository.NewListenersRepository()

	tradingService := service.NewTrading(ctx, listenerRepository, positionRepository, priceService, paymentService, repository.NewPgxTransactor(pool))
	tradingServer := handler.NewPrice(tradingService)

	ns := grpc.NewServer()
	pr.RegisterTradingServiceServer(ns, tradingServer)

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
