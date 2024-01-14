package jito

import (
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/gagliardetto/solana-go"
)

var jitoTipAccounts = []solana.PK{
	solana.MustPublicKeyFromBase58("ADuUkR4vqLUMWXxW9gh6D6L8pMSawimctcNZ5pGwDcEt"),
	solana.MustPublicKeyFromBase58("HFqU5x63VTqvQss8hp11i4wVV8bD44PvwucfZ2bU7gRe"),
	solana.MustPublicKeyFromBase58("DttWaMuVvTiduZRnguLF7jNxTgiMBZ1hyAumKUiL2KRL"),
	solana.MustPublicKeyFromBase58("Cw8CFyM9FkoMi7K7Crf6HNQqf4uEMzpKw6QNghXLvLkY"),
	solana.MustPublicKeyFromBase58("96gYZGLnJYVFmbjzopPSU6QiEV5fGqZNyN9nmNhvrZU5"),
	solana.MustPublicKeyFromBase58("3AVi9Tg9Uo68tJfuvoKvqKNWKkC5wPdSSdeBnizKZ6jT"),
	solana.MustPublicKeyFromBase58("ADaUMid9yfUytqMBgopwjb2DTLSokTSzL1zt6iGPaS49"),
	solana.MustPublicKeyFromBase58("DfXygSm4jCyNCybVYYK6DwvWqjKee8pbDmJGcLWNDXjh"),
}

const JITO_TIP_ACCOUNTS_COUNT = 8

var JitoTipLamports uint64

func init() {
	var err error
	JitoTipLamports, err = strconv.ParseUint(os.Getenv("JITO_TIP_LAMPORTS"), 10, 64)
	if err != nil || JitoTipLamports < 1000 {
		log.Fatalf("Error parsing JITO_TIP_LAMPORTS %v %v", JitoTipLamports, err)
	}

}

func GetRandomJitoTipAccount() solana.PK {
	return jitoTipAccounts[rand.Intn(8)]
}
