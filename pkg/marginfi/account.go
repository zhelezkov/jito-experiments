package marginfi

import (
	"encoding/binary"
	"jito-bot/pkg/fixed"

	"github.com/gagliardetto/solana-go"
)

type LendingAccount struct {
	Balances [MaxBalances]Balance
}

type MarginRequirementType uint

const (
	MarginRequirementTypeInitial MarginRequirementType = iota
	MarginRequirementTypeMaintenance
	MarginRequirementTypeEquity
)

type MarginfiAccount struct {
	// Group          solana.PublicKey
	Authority      solana.PublicKey
	LendingAccount LendingAccount
	// AccountFlags   uint64
}

func ParseMarginfiAccount(data []byte) *MarginfiAccount {
	authority := solana.PublicKey(data[40:72])
	lendingAccountData := data[72:]
	var lendingAccount LendingAccount

	for i := 0; i < 16; i++ {
		balanceData := lendingAccountData[i*96 : (i+1)*96]
		lendingAccount.Balances[i] = Balance{
			Active: balanceData[0] == 1,
			BankPK: solana.PublicKey(balanceData[1:33]),
			// padding 7 bytes
			AssetShares:          fixed.MustI80F48FromLittleEndian(balanceData[40:56]),
			LiabilityShares:      fixed.MustI80F48FromLittleEndian(balanceData[56:72]),
			EmissionsOutstanding: fixed.MustI80F48FromLittleEndian(balanceData[72:88]),
			LastUpdate:           binary.LittleEndian.Uint64(balanceData[88:96]),
		}
	}

	return &MarginfiAccount{
		Authority:      authority,
		LendingAccount: lendingAccount,
		// AccountFlags:   binary.LittleEndian.Uint64(data[1608:1616]),
	}
}

func (m *MarginfiAccount) CanBeLiquidated() bool {
	assets, liabilities := m.ComputeHealthComponents(nil, nil, MarginRequirementTypeMaintenance)
	return assets.LessThan(liabilities)
}

func (m *MarginfiAccount) ComputeHealthComponents(
	banks map[string]*Bank, oraclePrices map[string]*OraclePrice, req MarginRequirementType,
) (assets fixed.I80F48, liabilities fixed.I80F48) {
	for _, balance := range m.LendingAccount.Balances {
		// filter
		if !balance.Active {
			continue
		}
		// map
		bank := banks[balance.BankPK.String()]
		if bank == nil {
			panic("bank not found")
		}
		priceInfo := oraclePrices[bank.Group.String()]
		if priceInfo == nil {
			panic("priceInfo not found")
		}
		balanceAssets, balanceLiabilities := balance.GetUsdValueWithPriceBias(bank, priceInfo, req)
		// reduce
		assets = assets.Add(balanceAssets)
		liabilities = liabilities.Add(balanceLiabilities)
	}
	return
}
