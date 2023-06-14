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
	address2 "perun.network/perun-ckb-backend/wallet/address"
	"strconv"
)

type PaymentChannel struct {
	ch       *client.Channel
	currency channel.Asset
}

// newPaymentChannel creates a new payment channel.
func newPaymentChannel(ch *client.Channel, currency channel.Asset) *PaymentChannel {
	return &PaymentChannel{
		ch:       ch,
		currency: currency,
	}
}

func FormatState(c *PaymentChannel, state *channel.State, network types.Network) string {
	id := c.ch.ID()
	parties := c.ch.Params().Parts

	balA, _ := ShannonToCKByte(state.Allocation.Balance(0, c.currency)).Float64()
	balAStr := strconv.FormatFloat(balA, 'f', 4, 64)

	fstPartyPaymentAddr, _ := address2.AsParticipant(parties[0]).ToCKBAddress(network).Encode()
	sndPartyPaymentAddr, _ := address2.AsParticipant(parties[1]).ToCKBAddress(network).Encode()

	balB, _ := ShannonToCKByte(state.Allocation.Balance(1, c.currency)).Float64()
	balBStr := strconv.FormatFloat(balB, 'f', 4, 64)
	if len(parties) != 2 {
		log.Fatalf("invalid parties length: " + strconv.Itoa(len(parties)))
	}
	ret := fmt.Sprintf(
		"Channel ID: [green]%s[white]\nBalances:\n    %s: [green]%s[white] CKByte\n    %s: [green]%s[white] CKByte\nFinal: [green]%t[white]\nVersion: [green]%d[white]",
		hex.EncodeToString(id[:]),
		fstPartyPaymentAddr,
		balAStr,
		sndPartyPaymentAddr,
		balBStr,
		state.IsFinal,
		state.Version,
	)
	return ret
}

func (c PaymentChannel) State() *channel.State {
	return c.ch.State().Clone()
}

func (c PaymentChannel) SendPayment(amount float64) {
	// Transfer the given amount from us to peer.
	// Use UpdateBy to update the channel state.
	err := c.ch.Update(context.TODO(), func(state *channel.State) {
		shannonAmount := CKByteToShannon(big.NewFloat(amount))
		actor := c.ch.Idx()
		peer := 1 - actor
		state.Allocation.TransferBalance(actor, peer, c.currency, shannonAmount)
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
