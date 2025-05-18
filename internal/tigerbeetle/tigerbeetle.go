package tigerbeetle

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"unsafe"

	"math/big"
	"nova-backend-transaction-service/config"

	"github.com/google/uuid"
	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type TBClient interface {
	CreateAccount(ctx context.Context, accountID uuid.UUID, username string) (CAResult, error)
	TransferFunds(ctx context.Context, fromUser, toUser uuid.UUID, amount uint64) (TFResult, error)
	GetBalance(ctx context.Context, accountID uuid.UUID, fromTime, toTime uint64) (GBResult, error)
	GetTransfers(ctx context.Context, accountID uuid.UUID, fromTime, toTime uint64, limit bool) ([]GTResult, error)
}
type TBClientImpl struct {
	client tb.Client
}
type CAResult struct {
	AccountID string
	Timestamp string
}
type TFResult struct {
	TransferID string
	Timestamp  string
}
type GBResult struct {
	AccountID string
	Current   string
	Balances  []types.AccountBalance
	Timestamp string
}
type GTResult struct {
	TransferID   string
	FromUsername string
	ToUsername   string
	Amount       string
	Timestamp    string
}

func NewTBClient() TBClient {
	c := config.GetClient()
	log.Println("✅ TigerBeetle client initialized")
	return &TBClientImpl{client: c}
}

func (t *TBClientImpl) CreateAccount(ctx context.Context, accountID uuid.UUID, username string) (CAResult, error) {
	userCoded := (bytes16ToUint128(stringToBytes16(username)))
	clock := time.Now().UnixMicro()

	account := types.Account{
		ID:          bytes16ToUint128(accountID),
		Ledger:      1,
		UserData128: userCoded,
		Flags: types.AccountFlags{
			DebitsMustNotExceedCredits: true,
			History:                    true,
		}.ToUint16(),
		Timestamp: uint64(time.Now().UnixMicro()),
	}

	results, err := t.client.CreateAccounts([]types.Account{account})
	if err != nil {
		return CAResult{}, fmt.Errorf("tigerbeetle: create account error: %w", err)
	}
	if len(results) > 0 {
		return CAResult{}, fmt.Errorf("tigerbeetle: account creation failed with error code: %d", results[0])
	}
	return CAResult{
		AccountID: userCoded.String(),
		Timestamp: time.UnixMicro(clock).Format("2006-01-02 15:04"),
	}, nil
}

func (t *TBClientImpl) TransferFunds(ctx context.Context, fromUser, toUser uuid.UUID, amount uint64) (TFResult, error) {
	transferID := types.ID()
	clock := time.Now().UnixMicro()

	transfer := types.Transfer{
		ID:              transferID,
		DebitAccountID:  bytes16ToUint128(toUser),
		CreditAccountID: bytes16ToUint128(fromUser),
		Amount:          types.ToUint128(amount),
		Ledger:          1,
		Code:            1,
		Timestamp:       uint64(clock),
	}

	results, err := t.client.CreateTransfers([]types.Transfer{transfer})
	if err != nil {
		return TFResult{}, fmt.Errorf("tigerbeetle: create transfer error: %w", err)
	}
	if len(results) > 0 {
		return TFResult{}, fmt.Errorf("tigerbeetle: transfer creation failed with error code: %v", results[0].Result)
	}

	return TFResult{
		TransferID: transferID.String(),
		Timestamp:  time.UnixMicro(clock).Format("2006-01-02 15:04"),
	}, nil
}

func (t *TBClientImpl) GetBalance(ctx context.Context, accountID uuid.UUID, fromTime, toTime uint64) (GBResult, error) {
	if from, to := fromTime, toTime; from > to {
		return GBResult{}, errors.New("tigerbeetle: invalid time range")
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
		return GBResult{}, fmt.Errorf("tigerbeetle: create balance error: %v", err)
	}
	if len(balances) == 0 {
		return GBResult{
			AccountID: accountID.String(),
			Current:   "0",
			Balances:  []types.AccountBalance{},
			Timestamp: time.Now().Format("2006-01-02 15:04"),
		}, nil
	}
	currentBalance := new(big.Int)
	credits := balances[0].CreditsPosted.BigInt()
	debits := balances[0].DebitsPosted.BigInt()
	currentBalance.Sub(&credits, &debits)
	return GBResult{
		AccountID: accountID.String(),
		Current:   currentBalance.String(),
		Balances:  balances,
		Timestamp: time.Now().Format("2006-01-02 15:04"),
	}, nil
}

func (t *TBClientImpl) GetTransfers(ctx context.Context, accountID uuid.UUID, fromTime, toTime uint64, limit bool) ([]GTResult, error) {
	if from, to := fromTime, toTime; from > to {
		return []GTResult{}, errors.New("tigerbeetle: invalid time range")
	} else if to == 0 {
		to = uint64(time.Now().UnixMicro())
	}

	filter := types.AccountFilter{
		AccountID:    bytes16ToUint128(accountID),
		TimestampMin: fromTime,
		TimestampMax: toTime,
		Flags: types.AccountFilterFlags{
			Debits:   true,
			Credits:  true,
			Reversed: true,
		}.ToUint32(),
	}

	if limit {
		filter.Limit = 40
	}

	transfers, err := t.client.GetAccountTransfers(filter)
	if err != nil {
		return []GTResult{}, fmt.Errorf("tigerbeetle: create balance error: %v", err)
	}
	if len(transfers) == 0 {
		return []GTResult{}, nil
	}
	var movements []GTResult
	for _, transfer := range transfers {
		from, err := t.client.LookupAccounts([]types.Uint128{transfer.CreditAccountID})
		if err != nil {
			return []GTResult{}, fmt.Errorf("Unable to get usernames for this transfer: %v", err)
		}

		to, err := t.client.LookupAccounts([]types.Uint128{transfer.DebitAccountID})
		if err != nil {
			return []GTResult{}, fmt.Errorf("Unable to get usernames for this transfer: %v", err)
		}

		movements = append(movements, GTResult{
			TransferID:   transfer.ID.String(),
			FromUsername: string(bytes16ToString(uint128ToBytes16(from[0].UserData128))),
			ToUsername:   string(bytes16ToString(uint128ToBytes16(to[0].UserData128))),
			Amount:       transfer.Amount.String(),
			Timestamp:    time.Unix(0, int64(transfer.Timestamp)*1000).Format("2006-01-02 15:04"),
		})
	}
	return movements, nil
}

// Convert 16-byte array back to a string
func bytes16ToString(b [16]byte) string {
	return string(b[:])
}

// Convert Uint128 back to 16-byte array
func uint128ToBytes16(u types.Uint128) [16]byte {
	return *(*[16]byte)(unsafe.Pointer(&u))
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
