package marginfi

import (
	"encoding/binary"
	"jito-bot/pkg/fixed"
	"math"

	"github.com/gagliardetto/solana-go"
	"github.com/shabbyrobe/go-num"
)

type BankOperationalState uint8

const (
	BankOperationalStatePaused BankOperationalState = iota
	BankOperationalStateOperational
	BankOperationalStateReduceOnly
)

type OracleSetup uint8

const (
	OracleSetupNone OracleSetup = iota
	OracleSetupPyth
	OracleSetupSwitchboardV2
)

type BankConfig struct {
	AssetWeightInit  fixed.I80F48
	AssetWeightMaint fixed.I80F48

	LiabilityWeightInit  fixed.I80F48
	LiabilityWeightMaint fixed.I80F48

	DepositLimit uint64

	// InterestRateConfig struct {
	// 	OptimalUtilizationRate fixed.I80F48
	// 	PlateauInteresetRate   fixed.I80F48
	// 	MaxInterestRate        fixed.I80F48
	// 	InsuranceFeeFixedApr   fixed.I80F48
	// 	InsuranceIrFee         fixed.I80F48
	// 	ProtocolFixedFeeApr    fixed.I80F48
	// 	ProtocolIrFee          fixed.I80F48
	// 	// padding u128x8
	// } // size 128*7+128*8=1920/8=240 bytes
	OperationalState BankOperationalState

	OracleSetup OracleSetup
	OracleKeys  [5]solana.PublicKey
	// padding 6

	BorrowLimit uint64
}

type Bank struct {
	Mint         solana.PublicKey
	MintDecimals uint8
	Group        solana.PublicKey
	// padding 7
	AssetShareValue     fixed.I80F48
	LiabilityShareValue fixed.I80F48

	LiquidityVault              solana.PublicKey
	LiquidityVaultBump          uint8
	LiquidityVaultAuthorityBump uint8

	InsuranceVault              solana.PublicKey
	InsuranceVaultBump          uint8
	InsuranceVaultAuthorityBump uint8
	// padding 4
	CollectedInsuranceFeesOutstanding fixed.I80F48
	FeeVault                          solana.PublicKey
	FeeVaultBump                      uint8
	FeeVaultAuthorityBump             uint8
	// padding 6
	CollectedGroupFeesOutstanding fixed.I80F48
	TotalLiabilityShares          fixed.I80F48
	TotalAssetShares              fixed.I80F48

	LastUpdate int64

	Config BankConfig
}

func ParseBank(data []byte) *Bank {
	data = data[8:]
	// padding adds N bytes

	return &Bank{
		Mint:         solana.PublicKey(data[0:32]),
		MintDecimals: data[32],
		Group:        solana.PublicKey(data[33:65]),
		// padding 7
		AssetShareValue:     fixed.MustI80F48FromLittleEndian(data[72:88]),
		LiabilityShareValue: fixed.MustI80F48FromLittleEndian(data[88:104]),

		LiquidityVault:              solana.PublicKey(data[104:136]),
		LiquidityVaultBump:          data[136],
		LiquidityVaultAuthorityBump: data[137],
		InsuranceVault:              solana.PublicKey(data[138:170]),
		InsuranceVaultBump:          data[170],
		InsuranceVaultAuthorityBump: data[171],
		// padding 4
		CollectedInsuranceFeesOutstanding: fixed.MustI80F48FromLittleEndian(data[176:192]),
		FeeVault:                          solana.PublicKey(data[192:224]),
		FeeVaultBump:                      data[224],
		FeeVaultAuthorityBump:             data[225],
		// padding 6
		CollectedGroupFeesOutstanding: fixed.MustI80F48FromLittleEndian(data[232:248]),
		TotalLiabilityShares:          fixed.MustI80F48FromLittleEndian(data[248:264]),
		TotalAssetShares:              fixed.MustI80F48FromLittleEndian(data[264:280]),

		LastUpdate: int64(binary.LittleEndian.Uint64(data[280:288])),

		Config: BankConfig{
			AssetWeightInit:  fixed.MustI80F48FromLittleEndian(data[288:304]),
			AssetWeightMaint: fixed.MustI80F48FromLittleEndian(data[304:320]),

			LiabilityWeightInit:  fixed.MustI80F48FromLittleEndian(data[320:336]),
			LiabilityWeightMaint: fixed.MustI80F48FromLittleEndian(data[336:352]),

			DepositLimit: binary.LittleEndian.Uint64(data[352:360]),

			// InterestRateConfig padding 240

			OperationalState: BankOperationalState(data[600]),
			OracleSetup:      OracleSetup(data[601]),
			OracleKeys: [5]solana.PublicKey{
				solana.PublicKey(data[602:634]),
				solana.PublicKey(data[634:666]),
				solana.PublicKey(data[666:698]),
				solana.PublicKey(data[698:730]),
				solana.PublicKey(data[730:762]),
			},
			// padding 6
			BorrowLimit: binary.LittleEndian.Uint64(data[768:776]),
		},
	}
}

func (b *Bank) GetAssetQuantity(assetShares fixed.I80F48) fixed.I80F48 {
	return assetShares.Mul(b.AssetShareValue)
}

func (b *Bank) GetLiabilityQuantity(liabilityShares fixed.I80F48) fixed.I80F48 {
	return liabilityShares.Mul(b.LiabilityShareValue)
}

func (b *Bank) GetAssetShares(assetQuantity fixed.I80F48) fixed.I80F48 {
	return assetQuantity.Mul(b.AssetShareValue)
}

func (b *Bank) GetLiabilityShares(liabilityQuantity fixed.I80F48) fixed.I80F48 {
	return liabilityQuantity.Mul(b.LiabilityShareValue)
}

func (b *Bank) GetAssetWeight(reqType MarginRequirementType) fixed.I80F48 {
	switch reqType {
	case MarginRequirementTypeMaintenance:
		return b.Config.AssetWeightMaint
	case MarginRequirementTypeInitial:
	case MarginRequirementTypeEquity:
		panic("not implemented")
	}
	panic("unreachable")
}

func (b *Bank) GetLiabilityWeight(reqType MarginRequirementType) fixed.I80F48 {
	switch reqType {
	case MarginRequirementTypeInitial:
		return b.Config.LiabilityWeightInit
	case MarginRequirementTypeMaintenance:
		return b.Config.LiabilityWeightMaint
	case MarginRequirementTypeEquity:
		return fixed.I80F48{U128: num.U128FromUint(1)}
	}
	panic("unreachable")
}

func (b *Bank) ComputeAssetUsdValue(
	oraclePrice *OraclePrice, assetShares fixed.I80F48, reqType MarginRequirementType, bias PriceBias,
) fixed.I80F48 {
	assetQuantity := b.GetAssetQuantity(assetShares)
	assetWeight := b.GetAssetWeight(reqType)
	isWeighted := reqType == MarginRequirementTypeInitial
	return b.ComputeUsdValue(oraclePrice, assetQuantity, bias, isWeighted, assetWeight, true)
}

func (b *Bank) ComputeLiabilityUsdValue(
	oraclePrice *OraclePrice, liabilityShares fixed.I80F48, reqType MarginRequirementType, bias PriceBias,
) fixed.I80F48 {
	liabilityQuantity := b.GetLiabilityQuantity(liabilityShares)
	liabilityWeight := b.GetLiabilityWeight(reqType)
	isWeighted := reqType == MarginRequirementTypeInitial
	return b.ComputeUsdValue(oraclePrice, liabilityQuantity, bias, isWeighted, liabilityWeight, true)
}

func (b *Bank) ComputeUsdValue(
	oraclePrice *OraclePrice, quantity fixed.I80F48, bias PriceBias, isWeighted bool, weight fixed.I80F48, scaleToBase bool,
) (res fixed.I80F48) {
	price := GetPrice(oraclePrice, bias, isWeighted)
	res = quantity.Mul(price)
	if isWeighted {
		res = res.Mul(weight)
	}
	if scaleToBase {
		res = res.Div64(uint64(math.Pow10(int(b.MintDecimals))))
	}
	return
}
