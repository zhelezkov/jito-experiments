package marginfi

import (
	"encoding/binary"
	"github.com/gagliardetto/solana-go"
	"jito-bot/pkg/fixed"
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
		balanceData := lendingAccountData[i*104 : (i+1)*104]
		lendingAccount.Balances[i] = Balance{
			Active: balanceData[0] == 1,
			BankPK: solana.PublicKey(balanceData[1:33]),
			// padding 7 bytes
			AssetShares:          fixed.MustI80F48FromLittleEndian(balanceData[40:56]),
			LiabilityShares:      fixed.MustI80F48FromLittleEndian(balanceData[56:72]),
			EmissionsOutstanding: fixed.MustI80F48FromLittleEndian(balanceData[72:88]),
			LastUpdate:           binary.LittleEndian.Uint64(balanceData[88:96]),
			// padding 8 byte
		}
	}

	return &MarginfiAccount{
		Authority:      authority,
		LendingAccount: lendingAccount,
		// AccountFlags:   binary.LittleEndian.Uint64(data[1608:1616]),
	}
}

func (m *MarginfiAccount) CanBeLiquidated(client *Client) (canBeLiquidated bool, assets fixed.I80F48, liabilities fixed.I80F48) {
	assets, liabilities = m.ComputeHealthComponents(client.Banks, client.OraclePrices, MarginRequirementTypeMaintenance)
	canBeLiquidated = assets.LessThan(liabilities)
	return canBeLiquidated, assets, liabilities
}

func (m *MarginfiAccount) ComputeHealthComponents(
	banks BankMap, oraclePrices OraclePriceMap, req MarginRequirementType,
) (assets fixed.I80F48, liabilities fixed.I80F48) {
	assets = fixed.I80F48{}
	liabilities = fixed.I80F48{}
	for _, balance := range m.LendingAccount.Balances {
		// filter
		if !balance.Active {
			continue
		}
		// map
		bank := banks[balance.BankPK]
		if bank == nil {
			return fixed.I80F48{}, fixed.I80F48{}
			//panic("bank not found")
		}
		priceInfo := oraclePrices[balance.BankPK]
		if priceInfo == nil {
			return fixed.I80F48{}, fixed.I80F48{}
			//panic("priceInfo not found")
		}
		balanceAssets, balanceLiabilities := balance.GetUsdValueWithPriceBias(bank, priceInfo, req)
		// reduce
		assets = assets.Add(balanceAssets)
		liabilities = liabilities.Add(balanceLiabilities)
	}
	return
}
