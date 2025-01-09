package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/transaction"
)

type Block struct {
	BlockNumber  uint64                     `json:"block_number"`
	PrevHash     string                     `json:"prevHash"`
	Timestamp    int64                      `json:"timestamp"`
	Nonce        int                        `json:"nonce"`
	Transactions []*transaction.Transaction `json:"transactions"`
}

func NewBlock(prevHash string, nonce int, blockNumber uint64) *Block {
	block := new(Block)
	block.PrevHash = prevHash
	block.Timestamp = time.Now().UnixNano()
	block.Nonce = nonce
	block.Transactions = []*transaction.Transaction{}
	block.BlockNumber = blockNumber

	return block
}

func (b Block) ToJson() string {
	nb, err := json.Marshal(b)

	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (b Block) Hash() string {
	bs, _ := json.Marshal(b)
	sum := sha256.Sum256(bs)
	hexRep := hex.EncodeToString(sum[:32])
	formattedHexRep := constants.HEX_PREFIX + hexRep

	return formattedHexRep
}

func (b *Block) Mine(difficulty int) error {
	for {
		b.Timestamp = time.Now().UnixNano()
		guessHash := b.Hash()
		desiredHash := strings.Repeat("0", difficulty)
		ourSolutionHash := guessHash[2 : 2+difficulty]
		if ourSolutionHash == desiredHash {
			return nil
		}

		b.Nonce++
	}

}

func (b *Block) AddTransactionToTheBlock(txn *transaction.Transaction) error {
	if len(b.Transactions) == int(constants.TXN_PER_BLOCK_LIMIT) {
		return fmt.Errorf("transaction limit reached for block")
	}

	b.Transactions = append(b.Transactions, txn)
	return nil
}
