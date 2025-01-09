// utils/utils.go
package utils

import (
	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/blockchain"
)

func CompareBlocks(blocks1 []*block.Block, blocks2 []*block.Block) bool {
	if len(blocks1) != len(blocks2) {
		return false
	}
	for i := range blocks1 {
		if !CompareBlock(blocks1[i], blocks2[i]) {
			return false
		}
	}
	return true
}
func CompareBlock(block1 *block.Block, block2 *block.Block) bool {
	if block1.PrevHash != block2.PrevHash {
		return false
	}
	if block1.Hash() != block2.Hash() {
		return false
	}
	if block1.Nonce != block2.Nonce {
		return false
	}
	if block1.Timestamp != block2.Timestamp {
		return false
	}
	if len(block1.Transactions) != len(block2.Transactions) {
		return false
	}
	for i := range block1.Transactions {
		if block1.Transactions[i].From != block2.Transactions[i].From || block1.Transactions[i].To != block2.Transactions[i].To || block1.Transactions[i].Value != block2.Transactions[i].Value {
			return false
		}
	}
	return true
}
func CompareBlockchain(bc1 *blockchain.BlockchainStruct, bc2 *blockchain.BlockchainStruct) bool {
	if len(bc1.Blocks) != len(bc2.Blocks) {
		return false
	}

	for i := range bc1.Blocks {
		if !CompareBlock(bc1.Blocks[i], bc2.Blocks[i]) {
			return false
		}
	}
	return true
}
