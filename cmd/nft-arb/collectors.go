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
	"github.com/gagliardetto/solana-go/rpc"
)

var magicEdenCollectionToMinPrice map[string]uint64 = make(map[string]uint64)
var tensorCollectionToMaxSellPrice map[string]uint64 = make(map[string]uint64)

var dasApi helius.HeliusClient = helius.HeliusClient{
	Url: "https://mainnet.helius-rpc.com/?api-key=d3e02706-1eb1-4fcd-b11a-87d79aed0e5d",
}

var whitelistToCollection map[solana.PublicKey]*solana.PublicKey

func PreInit() {
	whitelists, err := tensor.FindAllWhitelists(solanaConnection)
	if err != nil {
		log.Fatalf("unable to find whitelists: %v", err)
	}

	slog.Info("found whitelists", "count", len(whitelists))

	whitelistToCollection = make(map[solana.PublicKey]*solana.PublicKey, len(whitelists))
	for _, wl := range whitelists {
		whitelistToCollection[wl.Address] = wl.Voc
	}
}

func CollectMEM2Listings() {
	slog.Info("collecting me v2 listings")
	meListings, err := me.FindAllM2SellterTradeStates(solanaConnection)
	if err != nil {
		log.Fatalf("unable to get ME listings: %v", err)
	}

	slog.Info("found listings", "count", len(meListings))

	listingsToCollectPrice := make([]*me.M2SellerTradeState, 0, len(meListings))
	for _, listing := range meListings {
		if !listing.PaymentMint.IsZero() {
			continue
		}
		listingsToCollectPrice = append(listingsToCollectPrice, listing)
	}

	slog.Info("filtered listings", "len(listingsToCollectPrice)", len(listingsToCollectPrice))
	assetIds := make([]string, 0, len(listingsToCollectPrice))
	for _, listing := range listingsToCollectPrice {
		assetIds = append(assetIds, listing.TokenMint.String())
	}
	CollectAssetMetadata(assetIds)

	for _, listing := range listingsToCollectPrice {
		collectionIdRes := rdb.Get(ctx, "asset:"+listing.TokenMint.String())
		if collectionIdRes.Err() != nil {
			// slog.Error("unable to get collectionId", "assetId", listing.TokenMint.String(), "err", collectionIdRes.Err())
			continue
		}
		collectionId := collectionIdRes.Val()
		royalty, err := rdb.HGet(ctx, "collection:"+collectionId, "royalty").Float64()
		if err != nil {
			slog.Error("unable to get royalty", "collection", collectionId, "err", err)
			continue
		}
		floatPrice := float64(listing.BuyerPrice)
		buyPrice := uint64(math.Round(floatPrice + floatPrice*royalty + floatPrice*me.TakerFee))
		// slog.Info("found me listing", "collection", collectionId, "assetId", listing.AssetId.String(), "buyPrice", buyPrice)
		if collectionPrice, ok := magicEdenCollectionToMinPrice[collectionId]; ok {
			if buyPrice < collectionPrice {
				magicEdenCollectionToMinPrice[collectionId] = buyPrice
			}
		} else {
			magicEdenCollectionToMinPrice[collectionId] = buyPrice
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

	listingsToCollectPrice := make([]*me.M3SellerTradeState, 0, len(meListings))
	for _, listing := range meListings {
		if !listing.PaymentMint.IsZero() {
			continue
		}
		listingsToCollectPrice = append(listingsToCollectPrice, listing)
	}

	slog.Info("filtered listings", "len(listingsToCollectPrice)", len(listingsToCollectPrice))

	assetIds := make([]string, 0, len(listingsToCollectPrice))
	for _, listing := range listingsToCollectPrice {
		assetIds = append(assetIds, listing.AssetId.String())
	}
	CollectAssetMetadata(assetIds)

	for _, listing := range listingsToCollectPrice {
		collectionIdRes := rdb.Get(ctx, "asset:"+listing.AssetId.String())
		if collectionIdRes.Err() != nil {
			// slog.Error("unable to get collectionId", "assetId", listing.AssetId.String(), "err", collectionIdRes.Err())
			continue
		}
		collectionId := collectionIdRes.Val()
		royalty, err := rdb.HGet(ctx, "collection:"+collectionId, "royalty").Float64()
		if err != nil {
			slog.Error("unable to get royalty", "collection", collectionId, "err", err)
			continue
		}
		floatPrice := float64(listing.BuyerPrice)
		buyPrice := uint64(math.Round(floatPrice + floatPrice*royalty + floatPrice*me.TakerFee))
		// slog.Info("found me listing", "collection", collectionId, "assetId", listing.AssetId.String(), "buyPrice", buyPrice)
		if collectionPrice, ok := magicEdenCollectionToMinPrice[collectionId]; ok {
			if buyPrice < collectionPrice {
				magicEdenCollectionToMinPrice[collectionId] = buyPrice
			}
		} else {
			magicEdenCollectionToMinPrice[collectionId] = buyPrice
		}
	}
}

func CollectAssetMetadata(assets []string) {
	isKnownSliceCmd := rdb.SMIsMember(ctx, "known_assets", assets)
	if isKnownSliceCmd.Err() != nil {
		log.Fatal("unable to check known assets", "err", isKnownSliceCmd.Err())
	}
	filteredAssets := make([]string, 0, len(assets))
	for i, ok := range isKnownSliceCmd.Val() {
		if ok {
			continue
		}
		filteredAssets = append(filteredAssets, assets[i])
	}
	assets = filteredAssets

	slog.Info("collecting asset metadata", "count", len(assets))

	for i := 0; i < len(assets); i += helius.MaxBatchSize {
		j := i + helius.MaxBatchSize
		if j > len(assets) {
			j = len(assets)
		}
		slog.Info("collecting asset metadata batch", "i", i, "j", j)
		// process
		batchedAssets := assets[i:j]

		assetsInfo, err := dasApi.GetAssetsBatch(batchedAssets)
		if err != nil {
			log.Fatalf("unable to get assets: %v", err)
		}

		result := assetsInfo.GetArray("result")
		for i, asset := range result {
			assetId := batchedAssets[i]
			grouping := asset.Get("grouping", "0")
			if grouping == nil {
				// slog.Info("no grouping", "asset", batchedAssets[i])
				continue
			}
			groupKey := string(grouping.GetStringBytes("group_key"))
			groupValue := string(grouping.GetStringBytes("group_value"))
			if groupKey != "collection" {
				// slog.Info("no collection", "asset", batchedAssets[i])
				continue
			}
			royalty := asset.GetFloat64("royalty", "percent")
			// cache royalty
			if err := rdb.HSet(ctx, "collection:"+groupValue, "royalty", royalty).Err(); err != nil {
				slog.Error("unable to cache royalty", "collection", groupValue, "err", err)
				continue
			}
			rdb.Set(ctx, "asset:"+assetId, groupValue, 0)
		}
	}
	rdb.SAdd(ctx, "known_assets", assets)
}

func CollectTensorCnftBids() {
	slog.Info("collecting tensor cnft bids")

	bids, err := tensor.FindAllCnftBidStates(solanaConnection)
	if err != nil {
		log.Fatalf("unable to get tensor bids: %v", err)
	}

	slog.Info("found tensor bids", "count", len(bids))

	filteredBids := make([]*tensor.CnftBidState, 0, len(bids))
	now := time.Now().Unix()
	for _, bid := range bids {
		if bid.Margin != nil && !bid.Margin.IsZero() {
			continue
		}
		if bid.Quantity != 1 || bid.FilledQuantity != 0 || !bid.Cosigner.IsZero() || bid.Expiry < now || bid.Currency != nil {
			continue
		}
		if _, ok := whitelistToCollection[bid.TargetId]; !ok {
			// slog.Info("no whitelist", "targetId", bid.TargetId.String())
			continue
		}
		filteredBids = append(filteredBids, bid)
	}

	slog.Info("filtered tensor bids", "count", len(filteredBids))

	for _, bid := range filteredBids {
		collectionId := whitelistToCollection[bid.TargetId].String()
		royalty, err := rdb.HGet(ctx, "collection:"+collectionId, "royalty").Float64()
		if err != nil {
			// slog.Error("unable to get royalty", "collection", collectionId, "err", err)
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

func CollectTensorNftBids() {
	slog.Info("collecting tensor nft bids")
	pools, err := tensor.FindAllTSwapPools(solanaConnection)
	if err != nil {
		log.Fatalf("unable to get tensor pools: %v", err)
	}

	slog.Info("found tensor nft pools", "count", len(pools))

	slog.Info("loading escrow accounts info")

	zero := uint64(0)
	escrowAccs, err := solanaConnection.GetProgramAccountsWithOpts(ctx, tensor.TSwapProgramAddress, &rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 0,
					Bytes:  []byte{0x4b, 0xc7, 0xfa, 0x3f, 0xf4, 0xd1, 0xeb, 0x78},
				},
			},
		},
		DataSlice: &rpc.DataSlice{
			Offset: &zero,
			Length: &zero,
		},
	})
	if err != nil {
		log.Fatalf("unable to get escrow accounts: %v", err)
	}

	slog.Info("found escrow accounts", "count", len(escrowAccs))

	escrowAccountToLamports := make(map[solana.PublicKey]uint64, len(escrowAccs))
	for _, escrowAcc := range escrowAccs {
		escrowAccountToLamports[escrowAcc.Pubkey] = escrowAcc.Account.Lamports
	}

	filteredPools := make([]*tensor.TSwapPool, 0, len(pools))
	for _, pool := range pools {
		if pool.Frozen != nil {
			continue
		}
		if pool.IsCosigned {
			continue
		}
		if pool.Margin != nil && !pool.Margin.IsZero() {
			continue
		}
		if pool.PoolConfig.PoolType != 0 || pool.PoolConfig.Delta != 0 {
			continue
		}
		if pool.TakerSellCount != 0 && pool.TakerSellCount <= pool.PoolStats.TakerSellCount {
			continue
		}
		if escrowAccountToLamports[pool.SolEscrow] < pool.PoolConfig.StartingPrice {
			continue
		}
		if _, ok := whitelistToCollection[pool.Whitelist]; !ok {
			// slog.Info("no whitelist", "targetId", bid.TargetId.String())
			continue
		}
		filteredPools = append(filteredPools, pool)
	}
	slog.Info("filtered tensor nft pools", "count", len(filteredPools))

	for _, pool := range filteredPools {
		collectionId := whitelistToCollection[pool.Whitelist].String()
		royalty, err := rdb.HGet(ctx, "collection:"+collectionId, "royalty").Float64()
		if err != nil {
			// slog.Error("unable to get royalty", "collection", collectionId, "err", err)
			continue
		}

		floatPrice := float64(pool.PoolConfig.StartingPrice)
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
