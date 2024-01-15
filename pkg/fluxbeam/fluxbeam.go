package fluxbeam

import "github.com/gagliardetto/solana-go"

var FLUXBEAM_PROGRAM_ADDRESS = solana.MustPublicKeyFromBase58("FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X")

func MakeSwapIx() {
	/*
		const dataLayout = struct([
			u8('instruction'),
			u64('amountIn'),
			u64('minimumAmountOut'),
		]);

		const data = Buffer.alloc(dataLayout.span);
		dataLayout.encode(
			{
				instruction: 1, // Swap instruction
				amountIn: amountIn,
				minimumAmountOut: minimumAmountOut,
			},
			data,
		);

		const keys = [
			{pubkey: tokenSwap, isSigner: false, isWritable: false},
			{pubkey: authority, isSigner: false, isWritable: false},
			{pubkey: userTransferAuthority, isSigner: true, isWritable: false},
			{pubkey: userSource, isSigner: false, isWritable: true},
			{pubkey: poolSource, isSigner: false, isWritable: true},
			{pubkey: poolDestination, isSigner: false, isWritable: true},
			{pubkey: userDestination, isSigner: false, isWritable: true},
			{pubkey: poolMint, isSigner: false, isWritable: true},
			{pubkey: feeAccount, isSigner: false, isWritable: true},
			{pubkey: sourceMint, isSigner: false, isWritable: false},
			{pubkey: destinationMint, isSigner: false, isWritable: false},
			{pubkey: sourceTokenProgramId, isSigner: false, isWritable: false},
			{pubkey: destinationTokenProgramId, isSigner: false, isWritable: false},
			{pubkey: poolTokenProgramId, isSigner: false, isWritable: false},
		];
		if (hostFeeAccount !== null) {
			keys.push({pubkey: hostFeeAccount, isSigner: false, isWritable: true});
		}
	*/

}
