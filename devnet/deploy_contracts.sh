#!/bin/bash

genesis=$(cat accounts/genesis-2.txt | awk '/testnet/ { count++; if (count == 2) print $2}')

cd contracts

if [ -d "migrations/dev" ]; then
  rm -rf "migrations/dev"
fi

capsule deploy --address $genesis --api "http://127.0.0.1:8114" --fee 0.01 <<< "Yes"
