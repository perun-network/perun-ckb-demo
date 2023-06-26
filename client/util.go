package client

import (
	"fmt"
	"math/big"
	gpchannel "perun.network/go-perun/channel"
	asset2 "perun.network/perun-demo-tui/asset"
)

// CKByteToShannon converts a given amount in CKByte to Shannon.
func CKByteToShannon(ckbyteAmount *big.Float) (shannonAmount *big.Int) {
	shannonPerCKByte := new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil)
	shannonPerCKByteFloat := new(big.Float).SetInt(shannonPerCKByte)
	shannonAmountFloat := new(big.Float).Mul(ckbyteAmount, shannonPerCKByteFloat)
	shannonAmount, _ = shannonAmountFloat.Int(nil)
	return shannonAmount
}

// ShannonToCKByte converts a given amount in Shannon to CKByte.
func ShannonToCKByte(shannonAmount *big.Int) (adaAmount *big.Float) {
	shannonPerCKByte := new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)
	shannonPerCKByteFloat := new(big.Float).SetInt(shannonPerCKByte)
	shannonAmountFloat := new(big.Float).SetInt(shannonAmount)
	return new(big.Float).Quo(shannonAmountFloat, shannonPerCKByteFloat)
}

func GetTuiAsset(tuiAssets []asset2.TUIAsset, a gpchannel.Asset) (asset2.TUIAsset, error) {
	for _, tuiAsset := range tuiAssets {
		if a.Equal(tuiAsset) {
			return tuiAsset, nil
		}
	}
	return asset2.TUIAsset{}, fmt.Errorf("tui asset not found")
}
