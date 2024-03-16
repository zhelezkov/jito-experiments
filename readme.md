## jito-experiments

This are my experiements with Solana Jito MEV
It contains a bunch of decoders for fluxbeam, marginfi, magic eden, pyth, raydium, tensor
Plus some utilities for I80F48 type implementation in Go.

Experiments itself were the followoing:
  - Raydium memetokens sniper(raydium-sniper) - Implementation that worked during January, and farmed a bunch of SOL
  - Nft Arbitrage bot(nft-arb) - Arbitrage bot for nft, from me <-> tensor. Not completed yet
  - Token sniper for fluxbeam(fluxbeam-sniper) - I haven't completed it, cause hypothesis failed. There was a few activity on fluxbeam, like 1 new token per an hour(?)
  - Liquidator for marginfi(liq) - I haven't completed it, cause wasn't able to capture big liquidations, only a small ones. Prob want to get back to it later

It was my first project in Go!