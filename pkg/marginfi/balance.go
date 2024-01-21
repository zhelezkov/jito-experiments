package marginfi

import (
	"github.com/gagliardetto/solana-go"
	"jito-bot/pkg/fixed"
)

const MaxBalances = 16

// Balance size 1 + 32 + 7 + 16 + 16 + 16 + 8 + 8 = 104
type Balance struct {
	Active bool
	BankPK solana.PublicKey
	// padding 7 bytes
	AssetShares          fixed.I80F48
	LiabilityShares      fixed.I80F48
	EmissionsOutstanding fixed.I80F48
	LastUpdate           uint64
}

func (b *Balance) GetUsdValueWithPriceBias(bank *Bank, oraclePrice *OraclePrice, req MarginRequirementType) (fixed.I80F48, fixed.I80F48) {
	assetsValue := bank.ComputeAssetUsdValue(oraclePrice, b.AssetShares, req, PriceBiasLowest)
	liabilitiesValue := bank.ComputeLiabilityUsdValue(oraclePrice, b.LiabilityShares, req, PriceBiasHighest)

	return assetsValue, liabilitiesValue
}
