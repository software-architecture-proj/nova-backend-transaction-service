package tigerbeetle

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"unsafe"

	"nova-backend-transaction-service/config"

	"github.com/google/uuid"
	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type TBClient interface {
	CreateAccount(ctx context.Context, accountID uuid.UUID) error
	TransferFunds(ctx context.Context, fromUser, toUser uuid.UUID, username string, amount uint64) error
	GetBalance(ctx context.Context, accountID uuid.UUID, fromTime, toTime uint64) (types.AccountBalance, error)
}

type TBClientImpl struct {
	client tb.Client
}

func NewTBClient() TBClient {
	c := config.GetClient()
	log.Println("✅ TigerBeetle client initialized")
	return &TBClientImpl{client: c}
}

func (t *TBClientImpl) CreateAccount(ctx context.Context, accountID uuid.UUID) error {
	account := types.Account{
		ID:     bytes16ToUint128(accountID),
		Ledger: 1,
		Flags: types.AccountFlags{
			DebitsMustNotExceedCredits: true,
			History:                    true,
		}.ToUint16(),
		Timestamp: uint64(time.Now().UnixMicro()),
	}

	results, err := t.client.CreateAccounts([]types.Account{account})
	if err != nil {
		return fmt.Errorf("tigerbeetle: create account error: %w", err)
	}
	if len(results) > 0 {
		return fmt.Errorf("tigerbeetle: account creation failed with error code: %d", results[0])
	}
	return nil
}

func (t *TBClientImpl) TransferFunds(ctx context.Context, fromUser, toUser uuid.UUID, username string, amount uint64) error {
	userCoded := (bytes16ToUint128(stringToBytes16(username)))

	transfer := types.Transfer{
		ID:              types.ID(),
		DebitAccountID:  bytes16ToUint128(fromUser),
		CreditAccountID: bytes16ToUint128(toUser),
		UserData128:     userCoded,
		Amount:          types.ToUint128(amount),
		Ledger:          1,
		Code:            1,
		Timestamp:       uint64(time.Now().UnixMicro()),
	}

	results, err := t.client.CreateTransfers([]types.Transfer{transfer})
	if err != nil {
		return fmt.Errorf("tigerbeetle: create transfer error: %w", err)
	}
	if len(results) > 0 {
		return fmt.Errorf("tigerbeetle: transfer creation failed with error code: %v", results[0].Result)
	}
	return nil
}

func (t *TBClientImpl) GetBalance(ctx context.Context, accountID uuid.UUID, fromTime, toTime uint64) (types.AccountBalance, error) {
	if from, to := fromTime, toTime; from > to {
		return types.AccountBalance{}, errors.New("tigerbeetle: invalid time range")
	} else if to == 0 {
		to = uint64(time.Now().UnixMicro())
	}
	balances, err := t.client.GetAccountBalances(
		types.AccountFilter{
			AccountID:    bytes16ToUint128(accountID),
			TimestampMin: fromTime,
			TimestampMax: toTime,
			Flags: types.AccountFilterFlags{
				Debits:   true,
				Credits:  true,
				Reversed: true,
			}.ToUint32(),
		})
	if err != nil {
		return types.AccountBalance{}, fmt.Errorf("tigerbeetle: create balance error: %v", err)
	}
	if len(balances) == 0 {
		return types.AccountBalance{}, nil
	}
	return balances[0], nil
}

// Convert 16-byte to Uint128
func bytes16ToUint128(id uuid.UUID) types.Uint128 {
	return types.BytesToUint128(id)
}

// Convert a string to a 16-byte array
func stringToBytes16(s string) [16]byte {
	var b [16]byte
	copy(b[:], s)
	return b
}

// Convert 16-byte array back to a string
func bytes16ToString(b [16]byte) string {
	return string(b[:])
}

// Convert Uint128 back to 16-byte array
func uint128ToBytes16(u types.Uint128) [16]byte {
	return *(*[16]byte)(unsafe.Pointer(&u))
}
