#!/bin/bash

genesis=$(cat accounts/genesis-1.txt | awk '/testnet/ { count++; if (count == 2) print $2}')

cd contracts

if [ -d "migrations/dev" ]; then
  rm -rf "migrations/dev"
fi

echo -e 'y\n\n' | capsule deploy --address $genesis
