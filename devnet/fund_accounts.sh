#!/bin/bash

genesis=$(cat accounts/genesis-1.txt | awk '/testnet/ && !found {print $2; found=1}')
alice=$(cat accounts/alice.txt | awk '/testnet/ && !found {print $2; found=1}')
bob=$(cat accounts/bob.txt | awk '/testnet/ && !found {print $2; found=1}')

echo -e '\n' | ckb-cli wallet transfer --from-account $genesis --to-address $alice --capacity 10000000000
sleep 2.0
echo -e '\n' | ckb-cli wallet transfer --from-account $genesis --to-address $bob --capacity 10000000000
