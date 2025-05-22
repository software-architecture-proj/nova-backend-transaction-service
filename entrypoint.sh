#!/bin/bash
set -e

./transactions &

sleep 10

echo -e "\n${GREEN}=== Creating default bank ===${NC}"

grpcurl -plaintext -proto transaction_service.proto \
    -d '{
        "user_id": "00000000-0000-0000-0000-000000000001",
        "username": "Bank 1",
        "bank": true
    }' \
    localhost:50051 transaction.TransactionService/Account

wait