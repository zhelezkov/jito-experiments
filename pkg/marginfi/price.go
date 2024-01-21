package marginfi

import (
	"jito-bot/pkg/fixed"
	"jito-bot/pkg/pyth"
)

type PriceBias uint

const (
	PriceBiasLowest PriceBias = iota
	PriceBiasNone
	PriceBiasHighest
)

type PriceWithConfidence struct {
	Price        fixed.I80F48
	Conf         fixed.I80F48
	LowestPrice  fixed.I80F48
	HighestPrice fixed.I80F48
}

type OraclePrice struct {
	PriceRealtime PriceWithConfidence
	//PriceWeighted PriceWithConfidence
}

var PythPriceConfIntervals = fixed.MustI80F48FromFloat64(pyth.PriceConfIntervals)

func ParseOraclePrice(setup OracleSetup, data []byte) *OraclePrice {
	if setup != OracleSetupPyth {
		panic("unsupported oracle setup")
	}

	pythPriceData := pyth.ParsePriceData(data)

	priceRealtime := fixed.MustI80F48FromFloat64(pythPriceData.Agg.Price)
	confRealtime := fixed.MustI80F48FromFloat64(pythPriceData.Agg.Conf)
	adjConfRealtime := confRealtime.Mul(PythPriceConfIntervals)

	//price := pythPriceData.
	return &OraclePrice{
		PriceRealtime: PriceWithConfidence{
			Price:        priceRealtime,
			Conf:         confRealtime,
			LowestPrice:  priceRealtime.Sub(adjConfRealtime),
			HighestPrice: priceRealtime.Add(adjConfRealtime),
		},
	}
	return nil
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
		panic("is weighted price not really implemented... need to do ema parsing from pyth")
		//return oraclePrice.PriceWeighted
	}
	return oraclePrice.PriceRealtime
}
