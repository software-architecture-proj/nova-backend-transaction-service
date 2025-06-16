package main

import (
	"log"
	"net"
	"nova-backend-transaction-service/service"

	pb "github.com/software-architecture-proj/nova-backend-common-protos/gen/go/transaction_service"
	"nova-backend-transaction-service/internal/tigerbeetle"

	"google.golang.org/grpc"
)

func main() {
	// Initialize TigerBeetle
	tbClient := tigerbeetle.NewTBClient()

	// Set up gRPC listener
	listener, err := net.Listen("tcp", ":50051") // Use any available port
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}
	log.Println("✅ Listening on port :50051")

	// Create new gRPC server
	grpcServer := grpc.NewServer()

	// Register the TransactionService
	pb.RegisterTransactionServiceServer(grpcServer, &service.TransactionService{
		TB: tbClient,
	})

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Failed to serve gRPC: %v", err)
	}
}
