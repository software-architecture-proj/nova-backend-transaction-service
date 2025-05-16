package model

import (
	"github.com/google/uuid"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Account struct {
	ID        types.Uint128 // Foreign Key ID from SQL
	Ledger    uint8         // Must be 1 for now
	Flags     uint16        // DebitsMustNotExceedCredits, History
	Timestamp uint64        // Unix timestamp
}

type Transfer struct {
	ID              types.Uint128 // TigerBeetle time-based ID
	DebitAccountID  uuid.UUID     // Foreign Key ID from SQL
	CreditAccountID uuid.UUID     // Foreign Key ID from SQL
	Amount          types.Uint128 // Amount in smallest unit
	Ledger          uint8         // Must be 1 for now
	Code            uint8         // Seems like a code of all possible transfers in app
	Timestamp       uint64        // Unix timestamp
}
