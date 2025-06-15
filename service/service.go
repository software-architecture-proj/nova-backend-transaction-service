package service

import (
	"context"
	"fmt"
	"time"

	pb "nova-backend-transaction-service/gen/go"
	"nova-backend-transaction-service/internal/tigerbeetle"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionService struct {
	pb.UnimplementedTransactionServiceServer
	TB tigerbeetle.TBClient
}

// Transfer handles a fund transfer between two accounts.
func (s *TransactionService) Transfer(ctx context.Context, req *pb.TransferFundsRequest) (*pb.TransferFundsResponse, error) {
	fromUUID, err := uuid.Parse(req.FromUserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from account ID: %v", err)
	}

	toUUID, err := uuid.Parse(req.ToUserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to account ID: %v", err)
	}

	if req.Amount == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be greater than 0")
	}

	res, err := s.TB.TransferFunds(ctx, fromUUID, toUUID, req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}

    // Create the notification producer
    producer, err := notification.NewProducer()
    if err != nil {
        log.Printf("Failed to create notification producer: %v", err)
        return
    }
    defer producer.Close()  // Always close the producer when done

    // Send the transaction notification
    err = producer.SendTransactionNotification(req.FromUserEmail, res.TransferID, req.Amount)
    if err != nil {
        log.Printf("Failed to send transaction notification: %v", err)
        return
    }

	return &pb.TransferFundsResponse{
		Success:    true,
		Message:    fmt.Sprintf("Transferred %d from %s to %s", req.Amount, req.FromUserId, req.ToUserId),
		TransferId: res.TransferID,
		Timestamp:  res.Timestamp,
	}, nil
}

// Account handles the creation of one account.
func (s *TransactionService) Account(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	AccountID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from account ID: %v", err)
	}

	res, err := s.TB.CreateAccount(ctx, AccountID, req.Username, req.Bank)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Account creation failed: %v", err)
	}

	return &pb.CreateAccountResponse{
		Success:   true,
		Message:   fmt.Sprintf("Account created for %s with TB_ID: %s", req.Username, res.AccountID),
		UserId:    res.AccountID,
		Timestamp: res.Timestamp,
	}, nil
}

// Balance handles the current and previous balance of an account.
func (s *TransactionService) Balance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	AccountID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from account ID: %v", err)
	}

	current, res, err := s.TB.GetBalance(ctx, AccountID, req.FromTime, req.ToTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Account creation failed: %v", err)
	}

	balances := make([]*pb.GBResult, 0, len(res))
	for _, b := range res {
		balances = append(balances, &pb.GBResult{
			Income:  b.Income.String(),
			Outcome: b.Outcome.String(),
		})
	}
	if len(balances) == 0 {
		fmt.Println("Balance is 0 for this account or it was not found.")
		return &pb.GetBalanceResponse{
			Success:   true,
			Message:   fmt.Sprintf("No balances found for %s", req.UserId),
			Timestamp: time.Now().Format("2006-01-02 15:04"),
			Balances:  []*pb.GBResult{},
		}, nil
	}

	return &pb.GetBalanceResponse{
		Success:   true,
		Message:   fmt.Sprintf("Balances for %s.", req.UserId),
		Current:   current.String(),
		Timestamp: time.Now().Format("2006-01-02 15:04"),
		Balances:  balances,
	}, nil
}

// Movements handles the history of transfers of each account.
func (s *TransactionService) Movements(ctx context.Context, req *pb.GetMovementsRequest) (*pb.GetMovementsResponse, error) {
	AccountID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from account ID: %v", err)
	}

	res, err := s.TB.GetTransfers(ctx, AccountID, req.FromTime, req.ToTime, req.Limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Account creation failed: %v", err)
	}
	movements := make([]*pb.GTResult, 0, len(res))
	for _, m := range res {
		movements = append(movements, &pb.GTResult{
			TransferId:   m.TransferID,
			FromUsername: m.FromUsername,
			ToUsername:   m.ToUsername,
			Amount:       m.Amount.String(),
			Timestamp:    m.Timestamp,
		})
	}

	if len(movements) == 0 {
		return &pb.GetMovementsResponse{
			Success:   true,
			Message:   fmt.Sprintf("No movements found for %s", req.UserId),
			Movements: []*pb.GTResult{},
		}, nil
	}

	return &pb.GetMovementsResponse{
		Success:   true,
		Message:   fmt.Sprintf("History of movements for: %s", req.UserId),
		Movements: movements,
	}, nil
}
