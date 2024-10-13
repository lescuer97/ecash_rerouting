package wallet

import (
	"fmt"

	w "github.com/elnosh/gonuts/wallet"
)

const WALLET_DB_1 = "WALLET_DB_1"
const WALLET_DB_2 = "WALLET_DB_2"
const ACTIVE_MINT = "ACTIVE_MINT"

func SetUpWallet(walletpath string, mint string) (*w.Wallet, error) {

	config := w.Config{
		WalletPath:     walletpath,
		CurrentMintURL: mint,
	}

	wallet, err := w.LoadWallet(config)
	if err != nil {
		return nil, fmt.Errorf("w.LoadWallet(config). %w", err)

	}
	return wallet, nil

}

func getBalance() {

}
