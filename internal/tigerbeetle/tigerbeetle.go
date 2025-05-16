package tigerbeetle

import (
	"context"

	"github.com/google/uuid"
)

type TBClient interface {
	CreateAccount(ctx context.Context, accountID uuid.UUID) error
	TransferFunds(ctx context.Context, from, to uuid.UUID, amount uint64) error
	GetBalance(ctx context.Context, accountID uuid.UUID) (uint64, error)
}
