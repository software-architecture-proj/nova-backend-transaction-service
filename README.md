# 🐯 Transaction Service - TigerBeetle Integration

This document describes how the `Transactions` service leverages [TigerBeetle](https://tigerbeetle.com) as a high-performance financial ledger. It outlines how accounts are structured, how values are encoded, and the rationale behind architectural decisions.

---

## 🧩 Overview

TigerBeetle is used to store and process accounts and transfers with extreme reliability and speed. This service does **not** store user information directly; instead, it syncs with a **separate SQL-based `User` service**.

Every TigerBeetle account represents a user from the SQL `User` service and is **fully encoded using unsigned integers** (`uint64`, `uint128`, `uint16`) to maximize performance and data integrity.

---

## 📦 Account Structure

All TigerBeetle accounts are created using the following structure:

```go
account := types.Account{
	ID:          iDCoded,
	Ledger:      1,
	Code:        1,
	UserData128: userCoded,
	UserData64:  uint64(clock),
	Flags: types.AccountFlags{
		History:                    true,
		CreditsMustNotExceedDebits: true,
	}.ToUint16(),
	Timestamp: 0,
}

```
---
## 🧾 Field Descriptions

| Field         | Type      | Description                                                                                                                                                                                      |
| ------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `ID`          | `uint64`  | Encoded UUID from the SQL `user` service. This acts as a foreign key to tie TigerBeetle accounts to user records.                                                                                |
| `Ledger`      | `uint64`  | Static value (`1`) representing a logical group of accounts. All user accounts belong to the same ledger.                                                                                        |
| `Code`        | `uint64`  | Static value (`1`) for optional account categorization. May be extended in the future.                                                                                                           |
| `UserData128` | `uint128` | Encoded representation of the user's username. Used for human-traceability and audit purposes.                                                                                                   |
| `UserData64`  | `uint64`  | Encoded representation of the user's creation timestamp (in local or Unix format).                                                                                                               |
| `Flags`       | `uint16`  | Bitfield controlling account behavior. Regular users are set with:<br>`History: true`,<br>`CreditsMustNotExceedDebits: true`.<br>Bank users have:<br>`DebitsMustNotExceedCredits: true` instead. |
| `Timestamp`   | `uint64`  | Set to `0` at creation. TigerBeetle maintains its own logical clock to ensure ordered operations.     

---

## 📦 Transfer Structure

Transfers represent the movement of funds between two TigerBeetle accounts. Each transfer is a single atomic operation logged in the TigerBeetle ledger.

```go
transfer := types.Transfer{
	ID:              transferID,
	DebitAccountID:  bytes16ToUint128(toUser),
	CreditAccountID: bytes16ToUint128(fromUser),
	UserData64:      uint64(clock),
	Amount:          types.ToUint128(amount),
	Ledger:          1,
	Code:            1,
}

```
---
## 🧾 Field Descriptions

| Field             | Type      | Description                                                                                                              |
| ----------------- | --------- | ------------------------------------------------------------------------------------------------------------------------ |
| `ID`              | `uint128` | Unique identifier for the transfer. This should be globally unique for idempotency and reconciliation.                   |
| `DebitAccountID`  | `uint128` | Account that will be debited. Typically corresponds to the **recipient** in internal logic (`toUser`).                   |
| `CreditAccountID` | `uint128` | Account that will be credited. Typically corresponds to the **sender** in internal logic (`fromUser`).                   |
| `UserData64`      | `uint64`  | Timestamp or other application-specific metadata (e.g., Unix time). Used for traceability or audit logs.                 |
| `Amount`          | `uint128` | Transfer amount in minor currency units (e.g., cents). Must be the same for both debit and credit sides.                 |
| `Ledger`          | `uint64`  | Static value (`1`) to match the ledger group used in accounts.                                                           |
| `Code`            | `uint64`  | Static or optional value to categorize the transfer type. Can be expanded for different transaction types in the future. |
