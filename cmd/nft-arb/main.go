package main

import (
	"context"
	"encoding/hex"
	"jito-bot/pkg/jito"
	mev "jito-bot/pkg/jito/gen"
	"log"
	"log/slog"
	"os"
	"strconv"

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

var TCOMP_PROGRAM_ADDRESS = solana.MustPublicKeyFromBase58("TCMPhJdwDryooaGtiocG1u3xcYbRpiJzb283XfCZsDp")

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

type CollectionInfo struct {
	BuyPrice  uint64
	Royalties float64
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

	// CollectMEM2Listings()
	CollectMEM3Listings()
	CollectTensorBids()

	for collectionId, meBuyPrice := range magicEdenCollectionToMinPrice {
		sellPrice, ok := tensorCollectionToMaxSellPrice[collectionId]
		if !ok {
			continue
		}
		if sellPrice > meBuyPrice {
			slog.Info("found arbitrage opportunity", "collection", collectionId, "meBuyPrice", meBuyPrice, "sellPrice", sellPrice)
		}
	}

	return

	searcher = mev.NewSearcherServiceClient(conn)

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

	mempoolSub, err := searcher.SubscribeMempool(ctx, &mev.MempoolSubscription{
		Regions: []string{"frankfurt", "amsterdam", "ny", "tokyo"},
		Msg: &mev.MempoolSubscription_ProgramV0Sub{
			ProgramV0Sub: &mev.ProgramSubscriptionV0{
				Programs: []string{"TCMPhJdwDryooaGtiocG1u3xcYbRpiJzb283XfCZsDp"},
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
				programId, err := tx.Message.Program(ix.ProgramIDIndex)
				if err != nil {
					log.Fatalf("unable to get program id: %v", err)
				}
				if programId != TCOMP_PROGRAM_ADDRESS {
					continue
				}

				ixDataDecoder := bin.NewBinDecoder(ix.Data)
				anchorHeader, err := ixDataDecoder.ReadNBytes(8)
				if err != nil {
					log.Fatalf("unable to read ixType: %v", err)
				}

				isListTx := hex.EncodeToString(anchorHeader) == "36aec14311298426"

				if !isListTx {
					break
				}
				slog.Info(tx.String())
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
