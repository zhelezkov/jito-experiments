package marginfi

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"math"
	"testing"
)

var connection = rpc.New("https://mainnet.helius-rpc.com/?api-key=d3e02706-1eb1-4fcd-b11a-87d79aed0e5d")

// var targetPK = solana.MustPublicKeyFromBase58("Bro5obNDbSZBeJpRYaAYCPdjgC7CBTe39W9KMryf7NBt")
var targetPK = solana.MustPublicKeyFromBase58("7ezantMELzNpUaBs7Dtrn49J1xogwYJkLQsppRGQoG4K")

func TestLiq(t *testing.T) {
	mfiClient, err := NewClient(connection)
	if err != nil {
		t.Fatal(err)
	}

	acc, err := connection.GetAccountInfo(context.Background(), targetPK)
	if err != nil {
		t.Fatal(err)
	}

	mfiAcc := ParseMarginfiAccount(acc.Value.Data.GetBinary())
	balanceSol := mfiAcc.LendingAccount.Balances[0]
	//balanceUSDC := mfiAcc.LendingAccount.Balances[1]

	spew.Dump(balanceSol)
	//spew.Dump(balanceUSDC)

	bankSol := mfiClient.Banks[balanceSol.BankPK]
	//bankUSDC := mfiClient.Banks[balanceUSDC.BankPK]

	b1 := bankSol.GetAssetQuantity(balanceSol.AssetShares).AsFloat64() * math.Pow10(-int(bankSol.MintDecimals))
	//b2 := bankUSDC.GetAssetQuantity(balanceUSDC.AssetShares).AsFloat64() * math.Pow10(-int(bankSol.MintDecimals))
	l1 := bankSol.GetLiabilityQuantity(balanceSol.LiabilityShares).AsFloat64() * math.Pow10(-int(bankSol.MintDecimals))
	//l2 := bankUSDC.GetLiabilityQuantity(balanceUSDC.LiabilityShares).AsFloat64() * math.Pow10(-int(bankSol.MintDecimals))
	spew.Dump(b1, l1)
	//spew.Dump(b1, b2, l1, l2)

	liquidated, assets, liabilities := mfiAcc.CanBeLiquidated(mfiClient)
	maxLiabilityPaydown := assets.Sub(liabilities)
	spew.Dump(liquidated, assets.AsFloat64(), liabilities.AsFloat64(), maxLiabilityPaydown.AsFloat64())
}
