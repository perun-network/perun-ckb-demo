package client

import (
	"context"
	"fmt"
	"log"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

// HandleProposal is the callback for incoming channel proposals.
func (p *PaymentClient) HandleProposal(prop client.ChannelProposal, r *client.ProposalResponder) {
	lcp, err := func() (*client.LedgerChannelProposalMsg, error) {
		// Ensure that we got a ledger channel proposal.
		lcp, ok := prop.(*client.LedgerChannelProposalMsg)
		if !ok {
			return nil, fmt.Errorf("invalid proposal type: %T", p)
		}

		// Check that we have the correct number of participants.
		if lcp.NumPeers() != 2 {
			return nil, fmt.Errorf("invalid number of participants: %d", lcp.NumPeers())
		}
		// Check that the channel has the expected assets and funding balances.
		const assetIdx = 0
		if err := channel.AssertAssetsEqual(lcp.InitBals.Assets, []channel.Asset{p.currency}); err != nil {
			return nil, fmt.Errorf("Invalid assets: %v\n", err)
		} else if lcp.FundingAgreement[assetIdx][0].Cmp(lcp.FundingAgreement[assetIdx][1]) != 0 {
			return nil, fmt.Errorf("invalid funding balance")
		}
		return lcp, nil
	}()
	if err != nil {
		_ = r.Reject(context.TODO(), err.Error())
	}

	// Create a channel accept message and send it.
	accept := lcp.Accept(
		p.WalletAddress(),        // The Account we use in the channel.
		client.WithRandomNonce(), // Our share of the channel nonce.
	)
	ch, err := r.Accept(context.TODO(), accept)
	if err != nil {
		log.Printf("Error accepting channel proposal: %v", err)
		return
	}

	//TODO: startWatching
	// Start the on-chain event watcher. It automatically handles disputes.
	p.startWatching(ch)

	// Store channel.
	p.channels <- newPaymentChannel(ch, p.currency)
	p.AcceptedChannel()
}

// HandleUpdate is the callback for incoming channel updates.
func (p *PaymentClient) HandleUpdate(cur *channel.State, next client.ChannelUpdate, r *client.UpdateResponder) {
	// We accept every update that increases our balance.
	err := func() error {
		err := channel.AssertAssetsEqual(cur.Assets, next.State.Assets)
		if err != nil {
			return fmt.Errorf("Invalid assets: %v", err)
		}

		receiverIdx := 1 - next.ActorIdx // This works because we are in a two-party channel.
		curBal := cur.Allocation.Balance(receiverIdx, p.currency)
		nextBal := next.State.Allocation.Balance(receiverIdx, p.currency)
		if nextBal.Cmp(curBal) < 0 {
			return fmt.Errorf("Invalid balance: %v", nextBal)
		}
		return nil
	}()
	if err != nil {
		_ = r.Reject(context.TODO(), err.Error())
	}

	// Send the acceptance message.
	err = r.Accept(context.TODO())
	if err != nil {
		panic(err)
	}
}

// HandleAdjudicatorEvent is the callback for smart contract events.
func (p *PaymentClient) HandleAdjudicatorEvent(e channel.AdjudicatorEvent) {
	log.Printf("Adjudicator event: type = %T, client = %v", e, p.Account)
}
