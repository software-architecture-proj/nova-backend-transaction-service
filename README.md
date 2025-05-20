# nova-backend-transactions

## Installation

Golang must be installed, machine should be running on Linux.

Open a terminal and run

```bash
  go mod init nova-backend-transaction-service
  go mod tidy
```

```bash
  ./data/tigerbeetle format --cluster=0 --replica=0 --replica-count=1 --development ./data/0_0.tigerbeetle
```
## Usage

Run
```bash
  ./data/tigerbeetle start --addresses=3000 --development ./data/0_0.tigerbeetle
```

Open another terminal and run
```bash
  go run server.go
```
    