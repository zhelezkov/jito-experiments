package main

import (
	"errors"
	mev "jito-bot/jito-mev"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	ata "github.com/gagliardetto/solana-go/programs/associated-token-account"
	budget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
)

var (
	RAYDIUM_PROGRAM_ADDRESS = solana.MustPublicKeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")
	MARKET_PROGRAM_ADDRESS  = solana.MustPublicKeyFromBase58("srmqPvymJeFKQ4zGQed1GFppgkRHL9kaELCbyksJtPX")
)

type RaydiumPoolKeys struct {
	Id         solana.PK
	BaseVault  solana.PK
	QuoteVault solana.PK

	MarketId         solana.PK
	MarketBaseVault  solana.PK
	MarketQuoteVault solana.PK
	MarketBids       solana.PK
	MarketAsks       solana.PK
	MarketEventQueue solana.PK
}

type MarketData struct {
	Id         solana.PK
	Bids       solana.PK
	Asks       solana.PK
	EventQueue solana.PK
	BaseVault  solana.PK
	QuoteVault solana.PK
}

const PACKET_ADDR = "0.0.0.0"

const SwapFixedInInstructionSize = 17 // bytes
var (
	authoritySeed    = []byte{97, 109, 109, 32, 97, 117, 116, 104, 111, 114, 105, 116, 121}
	openOrdersSeed   = []byte("open_order_associated_seed")
	targetOrdersSeed = []byte("target_associated_seed")
)

const COMPUTE_BUDGET_INSTRUCTIONS_COUNT = 2

var (
	computeBudgetInstructions = []solana.Instruction{
		budget.NewSetComputeUnitPriceInstruction(131_072).Build(),
		budget.NewSetComputeUnitLimitInstruction(65_536).Build(),
	}
)

type SwapSide int

const (
	SwapBuy SwapSide = iota
	SwapSell
)

func MakeRaydiumSwapBundle(wallet solana.PrivateKey, side SwapSide, tokenMint solana.PK, amount uint64, poolKeys *RaydiumPoolKeys, blockhash solana.Hash, bundleTxs []*solana.Transaction) (*mev.Bundle, error) {
	packets := make([]*mev.Packet, 0, len(bundleTxs)+2)
	for _, tx := range bundleTxs {
		txData, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		packets = append(packets, &mev.Packet{
			Data: txData,
			Meta: &mev.Meta{
				Port:        0,
				Addr:        PACKET_ADDR,
				SenderStake: 0,
				Size:        uint64(len(txData)),
			},
		})
	}

	swapTx, err := makeRaydiumSwapTx(wallet.PublicKey(), side, tokenMint, amount, poolKeys, blockhash)
	if err != nil {
		return nil, err
	}

	swapTx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if wallet.PublicKey().Equals(key) {
			return &wallet
		}
		return nil
	})
	swapTxData, err := swapTx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	packets = append(packets, &mev.Packet{
		Data: swapTxData,
		Meta: &mev.Meta{
			Port:        0,
			Addr:        PACKET_ADDR,
			SenderStake: 0,
			Size:        uint64(len(swapTxData)),
		}})

	return &mev.Bundle{
		Packets: packets,
	}, nil
}

func makeRaydiumSwapTx(wallet solana.PK, side SwapSide, tokenMint solana.PK, amountIn uint64, poolKeys *RaydiumPoolKeys, blockhash solana.Hash) (*solana.Transaction, error) {
	var tokenIn solana.PK
	var tokenOut solana.PK
	if side == SwapBuy {
		tokenIn = solana.WrappedSol
		tokenOut = tokenMint
	} else {
		tokenIn = tokenMint
		tokenOut = solana.WrappedSol
	}

	tokenAtaIn, _, err := solana.FindAssociatedTokenAddress(wallet, tokenIn)
	if err != nil {
		return nil, err
	}

	tokenAtaOut, _, err := solana.FindAssociatedTokenAddress(wallet, tokenOut)
	if err != nil {
		return nil, err
	}

	swapIx, err := makeSwapFixedInInstruction(wallet, tokenAtaIn, amountIn, tokenAtaOut, poolKeys)
	if err != nil {
		return nil, err
	}

	instructions := make([]solana.Instruction, 0, COMPUTE_BUDGET_INSTRUCTIONS_COUNT+3)
	instructions = append(instructions, computeBudgetInstructions...)

	if side == SwapBuy {
		ataIx := ata.NewCreateInstruction(wallet, wallet, tokenMint).Build()
		instructions = append(instructions, ataIx)
	}
	instructions = append(instructions, swapIx)

	// include jito tip only for buying, since it's bundled
	if side == SwapBuy {
		instructions = append(instructions, system.NewTransferInstruction(JITO_TIP_LAMPORTS, wallet, GetRandomJitoTipAccount()).Build())
	}

	return solana.NewTransaction(instructions, blockhash, solana.TransactionPayer(wallet))
}

func makeSwapFixedInInstruction(wallet solana.PK, tokenAccountIn solana.PK, amountIn uint64, tokenAccountOut solana.PK, poolKeys *RaydiumPoolKeys) (solana.Instruction, error) {
	authority, _, err := solana.FindProgramAddress([][]byte{authoritySeed}, RAYDIUM_PROGRAM_ADDRESS)
	if err != nil {
		return nil, err
	}
	openOrders, _, err := solana.FindProgramAddress([][]byte{RAYDIUM_PROGRAM_ADDRESS.Bytes(), poolKeys.MarketId.Bytes(), openOrdersSeed}, RAYDIUM_PROGRAM_ADDRESS)
	if err != nil {
		return nil, err
	}

	targetOrders, _, err := solana.FindProgramAddress([][]byte{RAYDIUM_PROGRAM_ADDRESS.Bytes(), poolKeys.MarketId.Bytes(), targetOrdersSeed}, RAYDIUM_PROGRAM_ADDRESS)
	if err != nil {
		return nil, err
	}

	marketAuthority, err := findMarketAuthority(poolKeys.MarketId)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice{
		// system
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
		// amm
		solana.NewAccountMeta(poolKeys.Id, true, false),
		solana.NewAccountMeta(authority, false, false),
		solana.NewAccountMeta(openOrders, true, false),
		// v4 only
		solana.NewAccountMeta(targetOrders, true, false),
		//
		solana.NewAccountMeta(poolKeys.BaseVault, true, false),
		solana.NewAccountMeta(poolKeys.QuoteVault, true, false),
		// serum КОГО ОН СЕРИТ ТО БЛЯТЬ?
		solana.NewAccountMeta(MARKET_PROGRAM_ADDRESS, false, false),
		solana.NewAccountMeta(poolKeys.MarketId, true, false),
		solana.NewAccountMeta(poolKeys.MarketBids, true, false),
		solana.NewAccountMeta(poolKeys.MarketAsks, true, false),
		solana.NewAccountMeta(poolKeys.MarketEventQueue, true, false),
		solana.NewAccountMeta(poolKeys.MarketBaseVault, true, false),
		solana.NewAccountMeta(poolKeys.MarketQuoteVault, true, false),
		solana.NewAccountMeta(marketAuthority, false, false),
		// user
		solana.NewAccountMeta(tokenAccountIn, true, false),
		solana.NewAccountMeta(tokenAccountOut, true, false),
		solana.NewAccountMeta(wallet, false, true),
	}

	data := make([]byte, SwapFixedInInstructionSize)
	data[0] = 9
	bin.LE.PutUint64(data[1:], amountIn)
	bin.LE.PutUint64(data[9:], 0)

	return solana.NewInstruction(RAYDIUM_PROGRAM_ADDRESS, accounts, data), nil
}

func findMarketAuthority(marketId solana.PK) (solana.PK, error) {
	seeds := [][]byte{marketId.Bytes()}
	nonce := byte(0)
	for nonce < 100 {
		seedsWithNonce := append(seeds, []byte{nonce, 0, 0, 0, 0, 0, 0, 0})
		pk, err := solana.CreateProgramAddress(seedsWithNonce, MARKET_PROGRAM_ADDRESS)
		if err == nil {
			return pk, err
		}
		nonce++
	}
	return solana.PK{}, errors.New("unable to find market authority")
}

func parseMarketAccount(data []byte) (*MarketData, error) {
	dec := bin.NewBinDecoder(data)
	dec.SkipBytes(5)

	dec.SkipBytes(8)

	marketId, err := dec.ReadBytes(32)
	if err != nil {
		return nil, err
	}

	dec.SkipBytes(8)

	// baseMint
	dec.SkipBytes(32)
	dec.SkipBytes(32)

	baseVault, err := dec.ReadBytes(32)
	if err != nil {
		return nil, err
	}
	dec.SkipBytes(8)
	dec.SkipBytes(8)

	qouteVault, err := dec.ReadBytes(32)
	if err != nil {
		return nil, err
	}
	dec.SkipBytes(8)
	dec.SkipBytes(8)

	dec.SkipBytes(8)

	dec.SkipBytes(32)
	eventQueue, err := dec.ReadBytes(32)
	if err != nil {
		return nil, err
	}

	bids, err := dec.ReadBytes(32)
	if err != nil {
		return nil, err
	}
	asks, err := dec.ReadBytes(32)
	if err != nil {
		return nil, err
	}

	return &MarketData{
		Id:         solana.PublicKey(marketId),
		Bids:       solana.PublicKey(bids),
		Asks:       solana.PublicKey(asks),
		EventQueue: solana.PublicKey(eventQueue),
		BaseVault:  solana.PublicKey(baseVault),
		QuoteVault: solana.PublicKey(qouteVault),
	}, nil
}
