// block/block.go
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"KNIRVCHAIN-MAIN/constants"
)

type Block struct {
	BlockNumber  uint64         `json:"block_number"`
	PrevHash     string         `json:"prevHash"`
	Timestamp    int64          `json:"timestamp"`
	Nonce        int            `json:"nonce"`
	Transactions []*Transaction `json:"transactions"`
	// HashVal      string                     `json:"hash"` // Removed HashVal
}

func NewBlock(prevHash string, nonce int, blockNumber uint64) *Block {
	block := new(Block)
	block.PrevHash = prevHash
	block.Timestamp = time.Now().UnixNano()
	block.Nonce = nonce
	block.Transactions = []*Transaction{}
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

//	func (b Block) MarshalJSON() ([]byte, error) {
//		log.Printf("Block Marshaling BlockNumber: %v, PrevHash: %v, Timestamp: %v, Nonce: %v, Transactions: %v", b.BlockNumber, b.PrevHash, b.Timestamp, b.Nonce, b.Transactions)
//
// type Alias Block
//
//	aux := struct {
//		Alias
//			Transactions []*Transaction `json:"transactions"`
//		}{
//		Alias:        (Alias)(b),
//		Transactions: b.Transactions,
//		}
//		return json.Marshal(aux)
//	}
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

func (b *Block) AddTransactionToTheBlock(txn *Transaction) error {
	// check if the txn verification is a success or a failure
	if txn.Status == constants.TXN_VERIFICATION_SUCCESS {
		txn.Status = constants.SUCCESS
	} else {
		txn.Status = constants.FAILED
	}

	b.Transactions = append(b.Transactions, txn)
	return nil
}
