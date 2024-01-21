package pyth

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func TestParse(t *testing.T) {
	rpc := rpc.New("https://mainnet.helius-rpc.com/?api-key=d3e02706-1eb1-4fcd-b11a-87d79aed0e5d")
	acc, err := rpc.GetMultipleAccounts(context.Background(), solana.MustPublicKeyFromBase58("H6ARHf6YXhGYeQfUzQNGk6rDNnLBQKrenN712K4AQJEG"))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", acc)

	price := ParsePriceData(acc.Value[0].Data.GetBinary())
	spew.Dump(price)
}
