package main

import (
	"context"
	"encoding/hex"
	"fmt"
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

	denominator, err := hex.DecodeString("c9a1e33e5cee6ff0")
	if err != nil {
		log.Fatalf("unable to decode denominator: %v", err)
	}

	bidState, err := solanaConnection.GetAccountInfo(ctx, solana.MustPublicKeyFromBase58("BXN23saFs9bpJCLbwK3j55jeKxGkNoD7MgTEzeta6CEC"))
	if err != nil {
		log.Fatalf("unable to get account info: %v", err)
	}

	binBidState := bidState.Bytes()

	// os.WriteFile("bidState.bin", binBidState, 0644)

	_ = binBidState

	decoder := bin.NewBinDecoder(binBidState)
	decoder.SkipBytes(8)                   // skip discriminator
	decoder.SkipBytes(1)                   // bump
	decoder.SkipBytes(32)                  // owner pk
	decoder.SkipBytes(32)                  // bidId ??
	decoder.SkipBytes(32)                  // target ? should be assetId ?????????
	targetId, err := decoder.ReadBytes(32) // targetId
	if err != nil {
		log.Fatalf("unable to read targetId: %v", err)
	}
	targetPk := solana.PublicKeyFromBytes(targetId)
	fmt.Println(targetPk.String())

	// str := hex.EncodeToString(binBidState[0:8])
	// fmt.Println(str)

	// data := [32]uint8{
	// 	// Offset 0x0000000A to 0x00000029
	// 	0xC4, 0x92, 0x87, 0xCC, 0x3E, 0x48, 0x73, 0x0A, 0x0A, 0xF7, 0xF2, 0x63,
	// 	0xD3, 0x1D, 0xF7, 0x3D, 0x55, 0x10, 0xFA, 0x29, 0x46, 0x5B, 0x4A, 0x27,
	// 	0xC9, 0xE3, 0x7E, 0xC2, 0xE5, 0xF8, 0xA2, 0xEE}
	// pk := solana.PublicKeyFromBytes(data[:])

	// fmt.Println(pk.String())
	// bin.NewBorshDecoder()

	zero := uint64(0)
	gpa, err := solanaConnection.GetProgramAccountsWithOpts(ctx, solana.MustPublicKeyFromBase58("TCMPhJdwDryooaGtiocG1u3xcYbRpiJzb283XfCZsDp"), &rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  denominator,
			},
		}},
		DataSlice: &rpc.DataSlice{
			Offset: &zero,
			Length: &zero,
		},
	})
	if err != nil {
		log.Fatalf("unable to get program accounts: %v", err)
	}

	fmt.Println(gpa)

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
