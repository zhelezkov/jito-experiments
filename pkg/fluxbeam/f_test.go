package fluxbeam

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func TestFxParse(t *testing.T) {
	connection := rpc.New("https://mainnet.helius-rpc.com/?api-key=d3e02706-1eb1-4fcd-b11a-87d79aed0e5d")

	v := uint64(0)
	txResp, err := connection.GetTransaction(context.Background(), solana.MustSignatureFromBase58("33a1JrzrQoSyMTAirbCwxZR4FtriWbgZFz49W3r1NRVGzGbxWVgJzxm7JstvXZiZ4X8Dr7P1vQqCfseYP7g7QxeB"), &rpc.GetTransactionOpts{
		MaxSupportedTransactionVersion: &v,
	})
	if err != nil {
		t.Fatal(err)
	}

	tx, err := txResp.Transaction.GetTransaction()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(tx.String())
}
