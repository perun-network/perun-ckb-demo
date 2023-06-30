package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"log"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/perun-ckb-backend/channel/asset"
	address2 "perun.network/perun-ckb-backend/wallet/address"
	asset2 "perun.network/perun-demo-tui/asset"
	"strconv"
)

type PaymentChannel struct {
	ch     *client.Channel
	assets []channel.Asset
}

// newPaymentChannel creates a new payment channel.
func newPaymentChannel(ch *client.Channel, assets []channel.Asset) *PaymentChannel {
	return &PaymentChannel{
		ch:     ch,
		assets: assets,
	}
}

func FormatState(c *PaymentChannel, state *channel.State, network types.Network, assetRegister asset2.Register) string {
	id := c.ch.ID()
	parties := c.ch.Params().Parts
	if len(parties) != 2 {
		log.Fatalf("invalid parties length: " + strconv.Itoa(len(parties)))
	}
	fstPartyPaymentAddr, _ := address2.AsParticipant(parties[0]).ToCKBAddress(network).Encode()
	sndPartyPaymentAddr, _ := address2.AsParticipant(parties[1]).ToCKBAddress(network).Encode()
	balAStrings := make([]string, len(c.assets))
	balBStrings := make([]string, len(c.assets))
	for i, a := range c.assets {
		if a.Equal(asset.CKBAsset) {
			balA, _ := ShannonToCKByte(state.Allocation.Balance(0, a)).Float64()
			balAStrings[i] = strconv.FormatFloat(balA, 'f', 2, 64)
			balB, _ := ShannonToCKByte(state.Allocation.Balance(1, a)).Float64()
			balBStrings[i] = strconv.FormatFloat(balB, 'f', 2, 64)
		} else {
			balAStrings[i] = state.Allocation.Balance(0, a).String()
			balBStrings[i] = state.Allocation.Balance(1, a).String()
		}
	}

	ret := fmt.Sprintf(
		"Channel ID: [green]%s[white]\n[red]Balances[white]:\n",
		hex.EncodeToString(id[:]),
	)
	ret += fmt.Sprintf("%s:\n", fstPartyPaymentAddr)
	for i, a := range c.assets {
		ret += fmt.Sprintf("    [green]%s[white] %s\n", balAStrings[i], assetRegister.GetName(a))
	}
	ret += fmt.Sprintf("%s:\n", sndPartyPaymentAddr)
	for i, a := range c.assets {
		ret += fmt.Sprintf("    [green]%s[white] %s\n", balBStrings[i], assetRegister.GetName(a))
	}
	ret += fmt.Sprintf("Final: [green]%t[white]\nVersion: [green]%d[white]", state.IsFinal, state.Version)
	return ret
}

func (c PaymentChannel) State() *channel.State {
	return c.ch.State().Clone()
}

func (c PaymentChannel) SendPayment(amounts map[channel.Asset]float64) {
	// Transfer the given amount from us to peer.
	// Use UpdateBy to update the channel state.
	err := c.ch.Update(context.TODO(), func(state *channel.State) {
		actor := c.ch.Idx()
		peer := 1 - actor
		for a, amount := range amounts {
			if amount < 0 {
				continue
			}
			if a.Equal(asset.CKBAsset) {
				shannonAmount := CKByteToShannon(big.NewFloat(amount))
				state.Allocation.TransferBalance(actor, peer, a, shannonAmount)
			} else {
				intAmount := new(big.Int).SetUint64(uint64(amount))
				state.Allocation.TransferBalance(actor, peer, a, intAmount)
			}
		}

	})
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err) // We panic on error to keep the code simple.
	}
}

// Settle settles the payment channel and withdraws the funds.
func (c PaymentChannel) Settle() {
	// Finalize the channel to enable fast settlement.
	if !c.ch.State().IsFinal {
		err := c.ch.Update(context.TODO(), func(state *channel.State) {
			state.IsFinal = true
		})
		if err != nil {
			panic(err)
		}
	}

	// Settle concludes the channel and withdraws the funds.
	err := c.ch.Settle(context.TODO(), false)
	if err != nil {
		panic(err)
	}

	// Close frees up channel resources.
	c.ch.Close()
}
