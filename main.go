package main

import (
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"log"
	"os"
	"perun.network/go-perun/wire"
	"perun.network/perun-ckb-backend/wallet"
	"perun.network/perun-ckb-demo/client"
	"perun.network/perun-ckb-demo/deployment"
	vc "perun.network/perun-demo-tui/client"
	"perun.network/perun-demo-tui/view"
)

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
	deployment.GetKey("./devnet/accounts/alice.pk")
	SetLogFile("demo.log")

	d, err := deployment.GetDeployment("./devnet/contracts/migrations/dev/", "./devnet/system_scripts")
	if err != nil {
		log.Fatalf("error getting deployment: %v", err)
	}

	w := wallet.NewEphemeralWallet()

	keyAlice, err := deployment.GetKey("./devnet/accounts/alice.pk")
	if err != nil {
		log.Fatalf("error getting alice's private key: %v", err)
	}
	keyBob, err := deployment.GetKey("./devnet/accounts/bob.pk")
	if err != nil {
		log.Fatalf("error getting bob's private key: %v", err)
	}
	aliceAccount := wallet.NewAccountFromPrivateKey(keyAlice)
	bobAccount := wallet.NewAccountFromPrivateKey(keyBob)

	err = w.AddAccount(aliceAccount)
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
		d,
		bus,
		rpcNodeURL,
		aliceAccount,
		*keyAlice,
		w,
	)
	if err != nil {
		log.Fatalf("error creating alice's client: %v", err)
	}
	bob, err := client.NewPaymentClient(
		"Bob",
		Network,
		d,
		bus,
		rpcNodeURL,
		bobAccount,
		*keyBob,
		w,
	)
	if err != nil {
		log.Fatalf("error creating bob's client: %v", err)
	}
	clients := []vc.DemoClient{alice, bob}
	_ = view.RunDemo("CKB Payment Channel Demo", clients)
}
