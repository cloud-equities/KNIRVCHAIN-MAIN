#!/bin/bash

# run chain app
go run main.go chain -port 5000 -miners_address testaddress -database_path ./knirv.db  &

# capture ID for other task operations.
CHAIN_PID=$!
echo "CHAIN STARTED. PID:$CHAIN_PID"
sleep 1

# run the vault command
go run main.go vault -port 8080 -node_address http://127.0.0.1:5000 -database_path ./knirv_test.db
sleep 1

echo "KILL CHAIN with PID $CHAIN_PID "
kill $CHAIN_PID
echo "KNIRV processes cleaned, please verify manually if required."