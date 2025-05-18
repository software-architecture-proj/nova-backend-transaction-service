package model

import (
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type AccountFinal struct {
	ID        types.Uint128 // Foreign Key ID from SQL
	Balances  types.AccountBalance
	Timestamp uint64 // Unix timestamp
}

type TransferFinal struct {
	ID                    types.Uint128 // TigerBeetle time-based ID
	DebitAccountUsername  string
	CreditAccountUsername string
	Amount                types.Uint128 // Amount in smallest unit
	Timestamp             uint64        // Unix timestamp
}
