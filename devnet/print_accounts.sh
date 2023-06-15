#!/bin/bash

for entry in ./accounts/*; do
  echo $entry
  account_id=$(cat $entry | awk '/lock_arg:/ { count++; print $2}')
  cat $entry
  echo $account_id
  echo -e '\n' | ckb-cli account export --lock-arg $account_id --extended-privkey-path ${entry%.*}.pk
  echo "------------------"
done
