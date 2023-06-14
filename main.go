package main

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"log"
	"os"
	"perun.network/go-perun/wire"
	"perun.network/perun-ckb-backend/backend"
	"perun.network/perun-ckb-backend/wallet"
	"perun.network/perun-ckb-demo/client"
	vc "perun.network/perun-demo-tui/client"
	"perun.network/perun-demo-tui/view"
)

var Deployment = backend.Deployment{
	Network:              types.NetworkTest,
	PCTSDep:              types.CellDep{},
	PCLSDep:              types.CellDep{},
	PFLSDep:              types.CellDep{},
	PCTSCodeHash:         types.Hash{},
	PCTSHashType:         "",
	PCLSCodeHash:         types.Hash{},
	PCLSHashType:         "",
	PFLSCodeHash:         types.Hash{},
	PFLSHashType:         "",
	PFLSMinCapacity:      0,
	DefaultLockScript:    types.Script{},
	DefaultLockScriptDep: types.CellDep{},
}

const (
	rpcNodeURL = "http://localhost:8114"
	Network    = types.NetworkTest
)

func SetLogFile(path string) {
	logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func main() {
	SetLogFile("demo.log")

	w := wallet.NewEphemeralWallet()

	keyAlice := secp256k1.PrivKeyFromBytes([]byte("alice")) // TODO: Add Alice's private key for demo!
	keyBob := secp256k1.PrivKeyFromBytes([]byte("bob"))     // TODO: Add Bob's private key for demo!
	aliceAccount := wallet.NewAccountFromPrivateKey(keyAlice)
	bobAccount := wallet.NewAccountFromPrivateKey(keyBob)

	err := w.AddAccount(aliceAccount)
	if err != nil {
		log.Fatalf("error adding alice's account: %v", err)
	}
	err = w.AddAccount(bobAccount)
	if err != nil {
		log.Fatalf("error adding bob's account: %v", err)
	}

	// Setup clients.
	log.Println("Setting up clients.")
	bus := wire.NewLocalBus() // Message bus used for off-chain communication.
	alice, err := client.NewPaymentClient(
		"Alice",
		Network,
		Deployment,
		bus,
		rpcNodeURL,
		aliceAccount,
		w,
	)
	if err != nil {
		log.Fatalf("error creating alice's client: %v", err)
	}
	bob, err := client.NewPaymentClient(
		"Bob",
		Network,
		Deployment,
		bus,
		rpcNodeURL,
		bobAccount,
		w,
	)
	if err != nil {
		log.Fatalf("error creating bob's client: %v", err)
	}
	clients := []vc.DemoClient{alice, bob}
	_ = view.RunDemo("CKB Payment Channel Demo", clients)
}
