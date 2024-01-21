package marginfi

import "jito-bot/pkg/fixed"

type PriceBias uint

const (
	PriceBiasLowest PriceBias = iota
	PriceBiasNone
	PriceBiasHighest
)

type PriceWithConfidence struct {
	Price        fixed.I80F48
	Confidence   fixed.I80F48
	LowestPrice  fixed.I80F48
	HighestPrice fixed.I80F48
}

type OraclePrice struct {
	PriceRealtime PriceWithConfidence
	PriceWighted  PriceWithConfidence
}

func GetPrice(oraclePrice *OraclePrice, bias PriceBias, isWeighted bool) (res fixed.I80F48) {
	price := GetPriceWithConfidence(oraclePrice, isWeighted)
	switch bias {
	case PriceBiasLowest:
		res = price.LowestPrice
	case PriceBiasNone:
		res = price.Price
	case PriceBiasHighest:
		res = price.HighestPrice
	}
	return
}

func GetPriceWithConfidence(oraclePrice *OraclePrice, isWeighted bool) PriceWithConfidence {
	if isWeighted {
		return oraclePrice.PriceWighted
	}
	return oraclePrice.PriceRealtime
}
