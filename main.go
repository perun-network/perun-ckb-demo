package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
	"perun.network/perun-ckb-backend/channel/asset"
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

type AssetRegister struct {
	getName  map[channel.Asset]string
	getAsset map[string]channel.Asset
	assets   []channel.Asset
}

func (a AssetRegister) GetAsset(name string) channel.Asset {
	return a.getAsset[name]
}

func (a AssetRegister) GetName(asset channel.Asset) string {
	return a.getName[asset]
}

func (a AssetRegister) GetAllAssets() []channel.Asset {
	return a.assets
}

func NewAssetRegister(assets []channel.Asset, names []string) (*AssetRegister, error) {
	assetRegister := &AssetRegister{
		getName:  make(map[channel.Asset]string),
		getAsset: make(map[string]channel.Asset),
		assets:   assets,
	}
	if len(assets) != len(names) {
		return nil, errors.New("length of assets and names must be equal")
	}
	for i, a := range assets {
		if a == nil {
			return nil, errors.New("asset cannot be nil")
		}
		if names[i] == "" {
			return nil, errors.New("name cannot be empty")
		}
		if assetRegister.getName[a] != "" {
			return nil, errors.New("duplicate asset")
		}
		if assetRegister.getAsset[names[i]] != nil {
			return nil, errors.New("duplicate name")
		}
		assetRegister.getName[a] = names[i]
		assetRegister.getAsset[names[i]] = a
	}
	return assetRegister, nil
}

func main() {
	SetLogFile("demo.log")
	d, sudtInfo, err := deployment.GetDeployment("./devnet/contracts/migrations/dev/", "./devnet/system_scripts")
	if err != nil {
		log.Fatalf("error getting deployment: %v", err)
	}
	sudtOwnerLockArg, err := parseSUDTOwnerLockArg("./devnet/accounts/sudt-owner-lock-hash.txt")
	if err != nil {
		log.Fatalf("error getting SUDT owner lock arg: %v", err)
	}
	sudtInfo.Script.Args = []byte(sudtOwnerLockArg)

	assetRegister, err := NewAssetRegister([]channel.Asset{asset.CKBAsset, &asset.SUDTAsset{
		TypeScript:  *sudtInfo.Script,
		MaxCapacity: 1_000 * 1_000 * 1_000 * 1_000,
	}}, []string{"CKBytes", "sudt"})
	if err != nil {
		log.Fatalf("error creating mapping: %v", err)
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
		assetRegister,
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
		assetRegister,
	)
	if err != nil {
		log.Fatalf("error creating bob's client: %v", err)
	}
	clients := []vc.DemoClient{alice, bob}
	_ = view.RunDemo("CKB Payment Channel Demo", clients, assetRegister)
}

func parseSUDTOwnerLockArg(pathToSUDTOwnerLockArg string) (string, error) {
	b, err := ioutil.ReadFile(pathToSUDTOwnerLockArg)
	if err != nil {
		return "", fmt.Errorf("reading sudt owner lock arg from file: %w", err)
	}
	sudtOwnerLockArg := string(b)
	if sudtOwnerLockArg == "" {
		return "", errors.New("sudt owner lock arg not found in file")
	}
	return sudtOwnerLockArg, nil
}
