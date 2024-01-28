package main

import (
	"jito-bot/pkg/helius"
	"jito-bot/pkg/me"
	"jito-bot/pkg/tensor"
	"log"
	"log/slog"
	"math"
	"time"

	"github.com/gagliardetto/solana-go"
)

var magicEdenCollectionToMinPrice map[string]uint64 = make(map[string]uint64)
var tensorCollectionToMaxSellPrice map[string]uint64 = make(map[string]uint64)

var dasApi helius.HeliusClient = helius.HeliusClient{
	Url: "https://mainnet.helius-rpc.com/?api-key=d3e02706-1eb1-4fcd-b11a-87d79aed0e5d",
}

func CollectMEM2Listings() {
	slog.Info("collecting me v2 listings")
	meListings, err := me.FindAllM2SellterTradeStates(solanaConnection)
	if err != nil {
		log.Fatalf("unable to get ME listings: %v", err)
	}

	slog.Info("found listings", "count", len(meListings))

	filteredListings := make([]*me.M2SellerTradeState, 0, len(meListings))
	for _, listing := range meListings {
		if !listing.PaymentMint.IsZero() {
			continue
		}
		filteredListings = append(filteredListings, listing)
	}

	slog.Info("filtered listings", "count", len(filteredListings))

	for i := 0; i < len(filteredListings); i += helius.MaxBatchSize {
		j := i + helius.MaxBatchSize
		if j > len(filteredListings) {
			j = len(filteredListings)
		}
		// process
		slice := filteredListings[i:j]
		assets := make([]string, 0, len(slice))
		for _, acc := range slice {
			assets = append(assets, acc.TokenMint.String())
		}

		assetsInfo, err := dasApi.GetAssetsBatch(assets)
		if err != nil {
			log.Fatalf("unable to get assets: %v", err)
		}

		result := assetsInfo.GetArray("result")
		for i, asset := range result {
			meAcc := slice[i]
			grouping := asset.Get("grouping", "0")
			if grouping == nil {
				slog.Info("no grouping", "asset", assets[i])
				continue
			}
			groupKey := string(grouping.GetStringBytes("group_key"))
			groupValue := string(grouping.GetStringBytes("group_value"))
			if groupKey != "collection" {
				slog.Info("no collection", "asset", assets[i])
				continue
			}
			royalty := asset.GetFloat64("royalty", "percent")
			// cache royalty
			rdb.HSet(ctx, "collection:"+groupValue, "royalty", royalty)

			floatPrice := float64(meAcc.BuyerPrice)
			buyPrice := uint64(math.Round(floatPrice + floatPrice*royalty + floatPrice*me.TakerFee))
			slog.Info("found me listing", "collection", groupValue, "assetId", meAcc.TokenMint.String(), "buyPrice", buyPrice)
			if collectionPrice, ok := magicEdenCollectionToMinPrice[groupValue]; ok {
				if buyPrice < collectionPrice {
					magicEdenCollectionToMinPrice[groupValue] = buyPrice
				}
			} else {
				magicEdenCollectionToMinPrice[groupValue] = buyPrice
			}
		}
	}
}

func CollectMEM3Listings() {
	slog.Info("collecting me v3 listings")
	meListings, err := me.FindAllM3SellterTradeStates(solanaConnection)
	if err != nil {
		log.Fatalf("unable to get ME listings: %v", err)
	}

	slog.Info("found listings", "count", len(meListings))

	filteredListings := make([]*me.M3SellerTradeState, 0, len(meListings))
	for _, listing := range meListings {
		if !listing.PaymentMint.IsZero() {
			continue
		}
		filteredListings = append(filteredListings, listing)
	}

	slog.Info("filtered listings", "count", len(filteredListings))

	for i := 0; i < len(filteredListings); i += helius.MaxBatchSize {
		j := i + helius.MaxBatchSize
		if j > len(filteredListings) {
			j = len(filteredListings)
		}
		// process
		slice := filteredListings[i:j]
		assets := make([]string, 0, len(slice))
		for _, acc := range slice {
			assets = append(assets, acc.AssetId.String())
		}

		assetsInfo, err := dasApi.GetAssetsBatch(assets)
		if err != nil {
			log.Fatalf("unable to get assets: %v", err)
		}

		result := assetsInfo.GetArray("result")
		for i, asset := range result {
			meAcc := slice[i]
			grouping := asset.Get("grouping", "0")
			if grouping == nil {
				slog.Info("no grouping", "asset", assets[i])
				continue
			}
			groupKey := string(grouping.GetStringBytes("group_key"))
			groupValue := string(grouping.GetStringBytes("group_value"))
			if groupKey != "collection" {
				slog.Info("no collection", "asset", assets[i])
				continue
			}
			royalty := asset.GetFloat64("royalty", "percent")
			// cache royalty
			rdb.HSet(ctx, "collection:"+groupValue, "royalty", royalty)

			floatPrice := float64(meAcc.BuyerPrice)
			buyPrice := uint64(math.Round(floatPrice + floatPrice*royalty + floatPrice*me.TakerFee))
			slog.Info("found me listing", "collection", groupValue, "assetId", meAcc.AssetId.String(), "buyPrice", buyPrice)
			if collectionPrice, ok := magicEdenCollectionToMinPrice[groupValue]; ok {
				if buyPrice < collectionPrice {
					magicEdenCollectionToMinPrice[groupValue] = buyPrice
				}
			} else {
				magicEdenCollectionToMinPrice[groupValue] = buyPrice
			}
		}
	}
}

func CollectTensorBids() {
	slog.Info("collecting tensor bids")
	whitelists, err := tensor.FindAllWhitelists(solanaConnection)
	if err != nil {
		log.Fatalf("unable to find whitelists: %v", err)
	}

	slog.Info("found whitelists", "count", len(whitelists))

	whitelistToCollection := make(map[solana.PublicKey]*solana.PublicKey, len(whitelists))
	for _, wl := range whitelists {
		whitelistToCollection[wl.Address] = wl.Voc
	}

	bids, err := tensor.FindAllBidStates(solanaConnection)
	if err != nil {
		log.Fatalf("unable to get tensor bids: %v", err)
	}

	slog.Info("found tensor bids", "count", len(bids))

	filteredBids := make([]*tensor.BidState, 0, len(bids))
	now := time.Now().Unix()
	for _, bid := range bids {
		if bid.Margin != nil && !bid.Margin.IsZero() {
			continue
		}
		if bid.Quantity != 1 || !bid.Cosigner.IsZero() || bid.Expiry < now || bid.Currency != nil {
			continue
		}
		if _, ok := whitelistToCollection[bid.TargetId]; !ok {
			slog.Info("no whitelist", "targetId", bid.TargetId.String())
			continue
		}
		filteredBids = append(filteredBids, bid)
	}

	slog.Info("filtered tensor bids", "count", len(filteredBids))

	for _, bid := range filteredBids {
		collectionId := whitelistToCollection[bid.TargetId].String()
		royalty, err := rdb.HGet(ctx, "collection:"+collectionId, "royalty").Float64()
		if err != nil {
			slog.Error("unable to get royalty", "collection", collectionId, "err", err)
			continue
		}

		floatPrice := float64(bid.Amount)
		sellPrice := uint64(math.Round(floatPrice - floatPrice*royalty - floatPrice*tensor.TakerFee))
		if collectionPrice, ok := tensorCollectionToMaxSellPrice[collectionId]; ok {
			if sellPrice > collectionPrice {
				tensorCollectionToMaxSellPrice[collectionId] = sellPrice
			}
		} else {
			tensorCollectionToMaxSellPrice[collectionId] = sellPrice
		}
	}
}
