package main

import (
	"context"
	"jito-bot/pkg/fixed"
	"jito-bot/pkg/jito"
	mev "jito-bot/pkg/jito/gen"
	"jito-bot/pkg/marginfi"
	"jito-bot/pkg/pyth"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/joho/godotenv/autoload"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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
	var err error

	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds)

	solanaConnection = rpc.New(os.Getenv("RPC_URL"))

	jitoAuthKey = solana.MustPrivateKeyFromBase58(os.Getenv("JITO_AUTH_PRIVATE_KEY"))
	blockEngineUrl = os.Getenv("JITO_BLOCK_ENGINE_URL")

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
		"jitoTipLamports", jito.JitoTipLamports)

	auth, err := jito.NewGrpcAuthHandler(blockEngineUrl, jitoAuthKey)
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

	marginfiClient, err := marginfi.NewClient(solanaConnection)
	if err != nil {
		log.Fatalf("unable to create marginfi client: %v", err)
	}

	gpa, err := solanaConnection.GetProgramAccountsWithOpts(ctx, solana.MustPublicKeyFromBase58("MFv2hWf31Z9kbCa1snEPYctwafyhdvnV7FZnsebVacA"), &rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 8,
				Bytes:  marginfi.GroupAddress.Bytes(),
			},
		}}})
	//{
	//	Memcmp: &rpc.RPCFilterMemcmp{
	//		Offset: 8 + 32,
	//		Bytes:  solana.MustPublicKeyFromBase58("2Hog8LpazH8kebxwvfWXRPdVu52pM2MGWHQbGVRzRsjp").Bytes(),
	//	},
	//}},

	tenBucks := fixed.MustI80F48FromFloat64(10)

	parsedAccounts := make([]*marginfi.MarginfiAccount, len(gpa))
	for i, gpaAcc := range gpa {
		parsedAccounts[i] = marginfi.ParseMarginfiAccount(gpaAcc.Account.Data.GetBinary())
	}

	now := time.Now()
	for _, account := range parsedAccounts {
		//account := marginfi.ParseMarginfiAccount(gpaAcc.Account.Data.GetBinary())
		canBeLiquidated, assets, liabilities := account.CanBeLiquidated(marginfiClient)
		if canBeLiquidated && liabilities.BiggerThanOrEqual(tenBucks) {
			maxLiabilityPaydown := assets.AsFloat64() - liabilities.AsFloat64()
			_ = maxLiabilityPaydown
			slog.Info("Account can be liquidated", "owner", account.Authority.String(), "health", maxLiabilityPaydown, "assets", assets.AsFloat64(), "liabilities", liabilities.AsFloat64())
		}
	}

	slog.Info("took", time.Since(now))

	//
	//spew.Dump(canBeLiquidated)

	//myAcc, err := solanaConnection.GetAccountInfo(ctx, solana.MustPublicKeyFromBase58("AbZoHRCWAV2RzWNTxssYZutmrQX7GMdc4RpTWZ8gtT6n"))
	//if err != nil {
	//	log.Fatalf("unable to get account info: %v", err)
	//}
	//
	//mAcc := marginfi.ParseMarginfiAccount(myAcc.Bytes())
	//spew.Dump(mAcc)
	//
	//bankAcc, err := solanaConnection.GetAccountInfo(ctx, solana.MustPublicKeyFromBase58("CCKtUs6Cgwo4aaQUmBPmyoApH2gUDErxNZCAntD6LYGh"))
	//if err != nil {
	//	log.Fatalf("unable to get account info: %v", err)
	//}
	//
	//bank := marginfi.ParseBank(bankAcc.Bytes())
	//spew.Dump(bank)
	//
	//myBalance := mAcc.LendingAccount.Balances[0]
	//
	//spew.Dump(bank.GetAssetQuantity(myBalance.AssetShares).AsFloat64())

	// balanceMy := fixed.I80F48{U128: num.MustU128FromString("287207082485166")}
	// spew.Dump(balanceMy.Float64())

	// bInt := big.Int{}

	// bankAssetShareValue := uint256.MustFromDecimal("287207082485166")
	// spew.Dump(bankAssetShareValue)
	// MULTIPLIER := uint256.MustFromDecimal("281474976710656")
	// rem := uint256.Int{}
	// div, _ := bankAssetShareValue.DivMod(bankAssetShareValue, MULTIPLIER, &rem)

	// res := div.Float64()
	// res += rem.Float64() / 281474976710656

	// spew.Dump(res)

	// spew.Dump(bankAssetShareValue.Lsh(bankAssetShareValue, 48).String())

	// balance := num.MustU128FromString("287207082485166")
	// MULTIPLIER := num.MustU128FromString("281474976710656")
	// div, mod := balance.QuoRem(MULTIPLIER)

	// res := div.AsFloat64()
	// res += mod.AsFloat64() / 281474976710656
	// spew.Dump(res)
	// balance.QuoRem()
	// mul := balance.Mul(bankBalance).AsBigFloat()
	// spew.Dump(new(big.Float).Quo(mul, big.NewFloat(2e48)).String())
	// spew.Dump(new(big.Int).Quo(mul, big.NewInt(0).set))

	// zero := uint64(0)
	// gpa, err := solanaConnection.GetProgramAccountsWithOpts(ctx, solana.MustPublicKeyFromBase58("MFv2hWf31Z9kbCa1snEPYctwafyhdvnV7FZnsebVacA"), &rpc.GetProgramAccountsOpts{
	// 	Filters: []rpc.RPCFilter{{
	// 		Memcmp: &rpc.RPCFilterMemcmp{
	// 			Offset: 8,
	// 			Bytes:  solana.MustPublicKeyFromBase58("4qp6Fx6tnZkY5Wropq9wUYgtFxXKwE6viZxFHg3rdAG8").Bytes(),
	// 		},
	// 	}},
	// 	DataSlice: &rpc.DataSlice{
	// 		Offset: &zero,
	// 		Length: &zero,
	// 	},
	// })
	// if err != nil {
	// 	log.Fatalf("unable to get program accounts: %v", err)
	// }

	// fmt.Println(len(gpa))

	return

	// pks := make([]string, len(gpa))
	// for i := 0; i < len(pks); i++ {
	// 	pks[i] = gpa[i].Pubkey.String()
	// }

	// accsChunk, err := solanaConnection.GetMultipleAccounts(ctx, pks...)
	// if err != nil {
	// 	log.Fatalf("unable to get multiple accounts: %v", err)
	// }

	// fmt.Println(accsChunk)

	// bundleResSub, err := searcher.SubscribeBundleResults(ctx, &mev.SubscribeBundleResultsRequest{})
	// if err != nil {
	// 	log.Fatalf("unable to subscribe: %v", err)
	// }
	// go func() {
	// 	for {
	// 		bundleRes, err := bundleResSub.Recv()
	// 		if err != nil {
	// 			log.Fatalf("unable to receive: %v", err)
	// 		}
	// 		slog.Info("bundle result:", "bundle", bundleRes)
	// 	}
	// }()

	// fmt.Println("subscribing...", len(pks))

	mempoolSub, err := searcher.SubscribeMempool(ctx, &mev.MempoolSubscription{
		// Regions: []string{"frankfurt", "amsterdam", "ny", "tokyo"},
		// Msg: &mev.MempoolSubscription_WlaV0Sub{
		// 	WlaV0Sub: &mev.WriteLockedAccountSubscriptionV0{
		// 		Accounts: []string{"H6ARHf6YXhGYeQfUzQNGk6rDNnLBQKrenN712K4AQJEG"},
		// 	},
		// },
		Msg: &mev.MempoolSubscription_ProgramV0Sub{
			ProgramV0Sub: &mev.ProgramSubscriptionV0{
				Programs: []string{"FsJ3A3u2vn5cTVofAjvy6y5kwABJAqYWpe4975bi2epH"},
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
			// fmt.Println(tx.String())
			for _, ix := range tx.Message.Instructions {
				if !pyth.IsUpdatePriceInstruction(ix.Data) {
					continue
				}
				priceAccount, err := tx.Message.Account(ix.Accounts[1])
				if err != nil {
					panic(err)
				}
				if priceAccount != solana.MustPublicKeyFromBase58("H6ARHf6YXhGYeQfUzQNGk6rDNnLBQKrenN712K4AQJEG") {
					continue
				}
				spew.Dump(pyth.ParseUpdatePriceInstruction(ix.Data))

				// 	programId, err := tx.Message.Program(ix.ProgramIDIndex)
				// 	if err != nil {
				// 		log.Fatalf("unable to get program id: %v", err)
				// 	}
				// 	if programId != TCOMP_PROGRAM_ADDRESS {
				// 		continue
				// 	}

				// 	ixDataDecoder := bin.NewBinDecoder(ix.Data)
				// 	anchorHeader, err := ixDataDecoder.ReadNBytes(8)
				// 	if err != nil {
				// 		log.Fatalf("unable to read ixType: %v", err)
				// 	}

				// 	isListTx := hex.EncodeToString(anchorHeader) == "36aec14311298426"

				// 	if !isListTx {
				// 		break
				// 	}
				// 	slog.Info(tx.String())

				// isPoolCreate := ixType == 0
				// if !isPoolCreate {
				// 	break
				// }
				// fmt.Println("got pool create instruction")
				// fmt.Println(tx.String())
				// if !isPoolCreate {
				// 	break
				// }
				// ixDataDecoder.SkipBytes(1) // skip nonce

				// openTimeRaw, err := ixDataDecoder.ReadUint64(binary.LittleEndian)
				// if err != nil {
				// 	log.Fatalf("unable to read openTime: %v", err)
				// }
				// openTime := int64(openTimeRaw)

				// slog.Info("create pool tx", "serverTime", notif.ServerSideTs.AsTime(), "expiration", notif.ExpirationTime.AsTime(), "poolOpenTime", time.Unix(openTime, 0).UTC())
				// if openTime > time.Now().Unix() {
				// 	break
				// }
			}
		}
	}
}

func walletSigner(key solana.PublicKey) *solana.PrivateKey {
	if wallet.PublicKey().Equals(key) {
		return &wallet
	}
	return nil
}
