package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"BankingService/account-service/internal/broker"
	delivery "BankingService/account-service/internal/delivery/grpc"
	"BankingService/account-service/internal/repository"
	pb "BankingService/pb/account"
)

func main() {
	db, err := repository.NewPostgresDB("127.0.0.1", "5433", "user", "password", "banking")
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	redisCache, err := repository.NewRedisCache("127.0.0.1", "6379")
	if err != nil {
		log.Fatalf("Redis error: %v", err)
	}

	repo := repository.NewAccountRepository(db)

	rmq, err := broker.NewRabbitMQ("amqp://guest:guest@127.0.0.1:5673/", repo, redisCache)
	if err != nil {
		log.Fatalf("RabbitMQ error: %v", err)
	}
	defer rmq.Close()

	rmq.ConsumeDeposits()
	rmq.ConsumeTransfers()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Listener error: %v", err)
	}

	grpcServer := grpc.NewServer()
	accountHandler := delivery.NewAccountHandler(repo, redisCache)
	pb.RegisterAccountServiceServer(grpcServer, accountHandler)

	log.Println("Account Service running on port 50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
