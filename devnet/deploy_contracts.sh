#!/bin/bash

ACCOUNTS_DIR="accounts"
PERUN_CONTRACTS_DIR="contracts"

genesis=$(cat $ACCOUNTS_DIR/genesis-2.txt | awk '/testnet/ { count++; if (count == 2) print $2}')

cd $PERUN_CONTRACTS_DIR

if [ -d "$PERUN_CONTRACTS_DIR/migrations/dev" ]; then
  rm -rf "$PERUN_CONTRACTS_DIR/migrations/dev"
fi

expect << EOF
spawn capsule deploy --address $genesis --api "http://127.0.0.1:8114" --fee 0.01
expect "Confirm deployment? (Yes/No)"
send "Yes\r"
expect "Password:"
send "\r"
expect eof
EOF
