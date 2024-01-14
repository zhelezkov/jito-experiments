package main

import (
	"context"
	"encoding/binary"
	"fmt"
	mev "jito-bot/jito-mev"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ctx = context.Background()

	solanaConnection *rpc.Client
	rdb              = redis.NewClient(&redis.Options{})
)

var (
	searcher       mev.SearcherServiceClient
	jitoAuthKey    solana.PrivateKey
	blockEngineUrl string
)

var (
	wallet              solana.PrivateKey
	tradeAmountLamports uint64
)

func init() {
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	solanaConnection = rpc.New(os.Getenv("RPC_URL"))

	jitoAuthKey = solana.MustPrivateKeyFromBase58(os.Getenv("JITO_AUTH_PRIVATE_KEY"))
	blockEngineUrl = os.Getenv("JITO_BLOCK_ENGINE_URL")
	JITO_TIP_LAMPORTS, err = strconv.ParseUint(os.Getenv("JITO_TIP_LAMPORTS"), 10, 64)
	if err != nil || JITO_TIP_LAMPORTS < 1000 {
		log.Fatal("Error parsing JITO_TIP_LAMPORTS", JITO_TIP_LAMPORTS, err)
	}

	wallet = solana.MustPrivateKeyFromBase58(os.Getenv("TRADER_PRIVATE_KEY"))
	tradeAmountLamports, err = strconv.ParseUint(os.Getenv("TRADE_AMOUNT_LAMPORTS"), 10, 64)
	if err != nil {
		log.Fatal("Error parsing TRADER_TRADE_AMOUNT_LAMPORTS", err)
	}
}

func main() {
	slog.Info("starting",
		"wallet", wallet.PublicKey().String(),
		"tradeAmountLamports", tradeAmountLamports,
		"blockEngineUrl", blockEngineUrl,
		"jitoTipLamports", JITO_TIP_LAMPORTS)

	auth, err := NewGrpcAuthHandler(blockEngineUrl, jitoAuthKey)
	if err != nil {
		log.Fatalf("unable to authenticate: %v", err)
	}

	conn, err := grpc.Dial(
		blockEngineUrl,
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithUnaryInterceptor(auth.UnaryInterceptor),
		grpc.WithStreamInterceptor(auth.StreamInterceptor),
	)
	if err != nil {
		log.Fatalf("problem with the server: %v", err)
	}
	defer conn.Close()

	searcher = mev.NewSearcherServiceClient(conn)

	bundleResSub, err := searcher.SubscribeBundleResults(ctx, &mev.SubscribeBundleResultsRequest{})
	if err != nil {
		log.Fatalf("unable to subscribe: %v", err)
	}
	go func() {
		for {
			bundleRes, err := bundleResSub.Recv()
			if err != nil {
				log.Fatalf("unable to receive: %v", err)
			}
			slog.Info("bundle result:", "bundle", bundleRes)
		}
	}()

	mempoolSub, err := searcher.SubscribeMempool(ctx, &mev.MempoolSubscription{
		Msg: &mev.MempoolSubscription_ProgramV0Sub{
			ProgramV0Sub: &mev.ProgramSubscriptionV0{
				Programs: []string{RAYDIUM_PROGRAM_ADDRESS.String()},
			},
		},
	})
	if err != nil {
		log.Fatalf("unable to subscribe: %v", err)
	}

	for {
		notif, err := mempoolSub.Recv()
		if err != nil {
			log.Fatalf("unable to receive: %v", err)
		}
		for _, msg := range notif.Transactions {
			tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(msg.Data))
			if err != nil {
				log.Fatalf("unable to decode transaction: %v", err)
			}
			for _, ix := range tx.Message.Instructions {
				programId, err := tx.Message.Program(ix.ProgramIDIndex)
				if err != nil {
					log.Fatalf("unable to get program id: %v", err)
				}
				if programId != RAYDIUM_PROGRAM_ADDRESS {
					continue
				}

				ixDataDecoder := bin.NewBinDecoder(ix.Data)
				ixType, err := ixDataDecoder.ReadByte()
				if err != nil {
					log.Fatalf("unable to read ixType: %v", err)
				}
				isPoolCreate := ixType == 1
				if !isPoolCreate {
					break
				}
				ixDataDecoder.SkipBytes(1) // skip nonce

				openTimeRaw, err := ixDataDecoder.ReadUint64(binary.LittleEndian)
				if err != nil {
					log.Fatalf("unable to read openTime: %v", err)
				}
				openTime := int64(openTimeRaw)

				slog.Info("create pool tx", "serverTime", notif.ServerSideTs.AsTime(), "expiration", notif.ExpirationTime.AsTime(), "poolOpenTime", time.Unix(openTime, 0).UTC())
				if openTime > time.Now().Unix() {
					break
				}

				// pool created LFG
				go handlePool(tx)
			}
		}
	}
}

func handlePool(tx *solana.Transaction) {
	var (
		poolId    solana.PK
		coinMint  solana.PK
		pcMint    solana.PK
		marketId  solana.PK
		coinVault solana.PK
		pcVault   solana.PK
	)

	keys := tx.Message.AccountKeys
	// tblKeys := tx.Message.GetAddressTableLookups().GetTableIDs()
	// resolutions := make(map[solana.PublicKey]solana.PublicKeySlice)
	// for _, key := range tblKeys {
	// 	fmt.Println("Getting table", key)

	// 	info, err := solanaConnection.GetAccountInfo(
	// 		ctx,
	// 		key,
	// 	)
	// 	if err != nil {
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }
	// 	// fmt.Println("got table " + key.String())

	// 	tableContent, err := lookup.DecodeAddressLookupTableState(info.GetBinary())
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println("table content:", spew.Sdump(tableContent))
	// 	fmt.Println("isActive", tableContent.IsActive())

	// 	resolutions[key] = tableContent.Addresses
	// }

	// tx.Message.SetAddressTables(resolutions)

	poolId = keys[2]
	coinMint = keys[17]
	pcMint = keys[14]
	marketId = keys[19]
	coinVault = keys[5]
	pcVault = keys[6]

	slog.Info("handling pool",
		"poolId", poolId,
		"coinMint", coinMint,
		"pcMint", pcMint,
		"marketId", marketId,
		"coinVault", coinVault,
		"pcVault", pcVault,
	)

	tokenMint := coinMint
	if tokenMint == solana.WrappedSol {
		tokenMint = pcMint
	}

	start := time.Now()
	market, err := findMarket(&marketId)
	slog.Info("find market took", "duration", time.Since(start))
	if err != nil {
		slog.Error("unable to find market", marketId.String(), err)
		return
	}

	start = time.Now()
	poolKeys := &RaydiumPoolKeys{
		Id:               poolId,
		BaseVault:        coinVault,
		QuoteVault:       pcVault,
		MarketId:         marketId,
		MarketBaseVault:  market.BaseVault,
		MarketQuoteVault: market.QuoteVault,
		MarketBids:       market.Bids,
		MarketAsks:       market.Asks,
		MarketEventQueue: market.EventQueue,
	}

	bundle, err := MakeRaydiumSwapBundle(wallet, SwapBuy, tokenMint, tradeAmountLamports, poolKeys, tx.Message.RecentBlockhash, []*solana.Transaction{tx})
	if err != nil {
		slog.Error("unable to make bundle", err)
		return
	}
	slog.Info("compose bundle took", "duration", time.Since(start))

	res, err := searcher.SendBundle(ctx, &mev.SendBundleRequest{
		Bundle: bundle,
	})
	if err != nil {
		slog.Error("unable to send bundle", err)
		return
	}

	slog.Info("bundle sent", "UUID", res.Uuid)

	// start selling after 2 sec
	time.Sleep(2 * time.Second)

	go tryDoSell(tokenMint, poolKeys)
}

func tryDoSell(tokenMint solana.PK, poolKeys *RaydiumPoolKeys) {
	slog.Info("trying to sell", "tokenMint", tokenMint.String())

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	tokenAta, _, err := solana.FindAssociatedTokenAddress(wallet.PublicKey(), tokenMint)
	if err != nil {
		slog.Error("unable to find associated token address", err)
		return
	}
loop:
	balance := "0"
	maxBalanceRetryAttempts := 40 // 40 attempts with 200ms sleep = 8 seconds
	for maxBalanceRetryAttempts > 0 {
		rpcBalance, err := solanaConnection.GetTokenAccountBalance(ctx, tokenAta, rpc.CommitmentConfirmed)
		if err == nil {
			balance = rpcBalance.Value.Amount
			break
		}
		maxBalanceRetryAttempts -= 1
		<-ticker.C
	}

	if balance == "0" {
		slog.Info("zero balance, nothing to sell :(")
		return
	}

	slog.Info("got balance", "balance", balance)

	amountToSell, err := strconv.ParseUint(balance, 10, 64)
	if err != nil {
		slog.Error("unable to parse balance")
		return
	}

	// var txId solana.Signature
	maxTxSendAttempts := 20 // 20 attempts with 200ms sleep = 4 seconds
	for maxTxSendAttempts > 0 {
		blockhash, err := solanaConnection.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
		if err != nil {
			slog.Error("unable to get blockhash", err)
			return
		}
		tx, err := makeRaydiumSwapTx(wallet.PublicKey(), SwapSell, tokenMint, amountToSell, poolKeys, blockhash.Value.Blockhash)
		if err != nil {
			slog.Error("unable to make tx", err)
			return
		}
		_, err = tx.Sign(walletSigner)
		if err != nil {
			slog.Error("unable to sign tx", err)
			return
		}
		txId, err := solanaConnection.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{SkipPreflight: true})
		if err != nil {
			slog.Error("unable to send tx", err)
			log.Println(tx.String())
		} else {
			slog.Info("sent sell tx", "sig", txId)
		}
		maxTxSendAttempts -= 1
		<-ticker.C
	}

	goto loop
}

func walletSigner(key solana.PublicKey) *solana.PrivateKey {
	if wallet.PublicKey().Equals(key) {
		return &wallet
	}
	return nil
}

func findMarket(marketId *solana.PK) (*MarketData, error) {
	marketFromCache, err := rdb.HGetAll(ctx, "market:"+marketId.String()).Result()
	if err != nil {
		return nil, err
	}
	if len(marketFromCache) > 0 {
		return &MarketData{
			Id:         solana.MustPublicKeyFromBase58(marketFromCache["id"]),
			Bids:       solana.MustPublicKeyFromBase58(marketFromCache["bids"]),
			Asks:       solana.MustPublicKeyFromBase58(marketFromCache["asks"]),
			EventQueue: solana.MustPublicKeyFromBase58(marketFromCache["eventQueue"]),
			BaseVault:  solana.MustPublicKeyFromBase58(marketFromCache["baseVault"]),
			QuoteVault: solana.MustPublicKeyFromBase58(marketFromCache["quoteVault"]),
		}, nil
	}

	slog.Info("market not found in cache, fetching from blockchain")

	acc, err := solanaConnection.GetAccountInfo(ctx, *marketId)
	if err != nil {
		return nil, err
	}

	market, err := parseMarketAccount(acc.Value.Data.GetBinary())
	if err != nil || market.Id != *marketId {
		return nil, fmt.Errorf("market id mismatch")
	}

	return market, nil
}
