package client

import (
	"context"
	"github.com/nervosnetwork/ckb-sdk-go/v2/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"log"
	"math"
	"math/big"
	"perun.network/perun-ckb-backend/wallet/address"
	"strconv"
	"time"
)

func (p *PaymentClient) PollBalances() {
	defer log.Println("PollBalances: stopped")
	pollingInterval := time.Second
	searchKey := &indexer.SearchKey{
		Script:           address.AsParticipant(p.Account.Address()).PaymentScript,
		ScriptType:       types.ScriptTypeLock,
		ScriptSearchMode: types.ScriptSearchModeExact,
		Filter:           nil,
		WithData:         false,
	}
	log.Println("PollBalances")
	updateBalance := func() {
		ctx, _ := context.WithTimeout(context.Background(), pollingInterval)

		cells, err := p.rpcClient.GetCells(ctx, searchKey, indexer.SearchOrderDesc, math.MaxUint32, "")
		if err != nil {
			log.Println("balance poll error: ", err)
			return
		}
		balance := big.NewInt(0)
		for _, cell := range cells.Objects {
			balance = new(big.Int).Add(balance, new(big.Int).SetUint64(cell.Output.Capacity))
		}

		p.balanceMutex.Lock()
		if balance.Cmp(p.balance) != 0 {
			p.balance = balance
			bal := p.balance.Int64()
			p.balanceMutex.Unlock()
			p.NotifyAllBalance(bal) // TODO: Update demo tui to allow for big.Int balances
		} else {
			p.balanceMutex.Unlock()
		}
	}
	// Poll the balance every 5 seconds.
	for {
		updateBalance()
		time.Sleep(pollingInterval)
	}
}

func FormatBalance(bal *big.Int) string {
	log.Printf("balance: %s", bal.String())
	balCKByte, _ := ShannonToCKByte(bal).Float64()
	return strconv.FormatFloat(balCKByte, 'f', 2, 64) + " CKByte"
}
