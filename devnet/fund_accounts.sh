#!/bin/bash

genesis=$(cat accounts/genesis-1.txt | awk '/testnet/ && !found {print $2; found=1}')
alice=$(cat accounts/alice.txt | awk '/testnet/ && !found {print $2; found=1}')
bob=$(cat accounts/bob.txt | awk '/testnet/ && !found {print $2; found=1}')
me=ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr0vthywaa50xwsqtm6mjd6xdfmyq3v69n2749jeddxakpgps7xqf24

genesis_tx_hash=$(ckb-cli wallet get-live-cells --address $genesis | awk '/tx_hash/ {print $2}')
genesis_tx_index=$(ckb-cli wallet get-live-cells --address $genesis | awk '/output_index/ && !found {print $2; found=1}')
genesis_tx_amount=$(ckb-cli wallet get-live-cells --address $genesis | awk '/capacity/ {print $3}')
FUNDINGTX="fundingtx.json"
FUNDING_AMOUNT=1000
CHANGE_AMOUNT=$(python -c "print(\"{:.8f}\".format($genesis_tx_amount - 2.0 * 10.0 * $FUNDING_AMOUNT - 1.0))")

add_output() {
  ckb-cli tx add-output --tx-file $FUNDINGTX --to-sighash-address $1 --capacity $2
}

ckb-cli tx init --tx-file $FUNDINGTX


fund_alice() {
  for ((i=1; i <= 10; i++)); do
    add_output $alice $FUNDING_AMOUNT
  done
}

fund_bob() {
  for ((i=1; i <= 10; i++)); do
    add_output $bob $FUNDING_AMOUNT
  done
}

fund_me() {
  for ((i=1; i <= 20; i++)); do
    add_output $me $FUNDING_AMOUNT
  done
}

if [ $# -eq 0 ]; then
  fund_alice
  fund_bob
else
  case $1 in
    alice)
      fund_alice
      ;;
    bob)
      fund_bob
      ;;
    me)
      fund_me
      ;;
    *)
      echo "Invalid argument"
      echo "Usage: fund_accounts.sh [alice|bob|me]"
      exit 1
      ;;
  esac
fi

ckb-cli tx add-output --tx-file $FUNDINGTX --to-sighash-address $genesis --capacity $CHANGE_AMOUNT
ckb-cli tx add-input --tx-file $FUNDINGTX --tx-hash $genesis_tx_hash --index $genesis_tx_index
ckb-cli tx sign-inputs --add-signatures --tx-file $FUNDINGTX --from-account $genesis
ckb-cli tx send --tx-file $FUNDINGTX
ckb-cli tx info --tx-file $FUNDINGTX
rm $FUNDINGTX
