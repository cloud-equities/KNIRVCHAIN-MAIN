package chain

import (
	"constants"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

type Block struct {
	BlockNumber  uint64         `json:"block_number"`
	PrevHash     []byte         `json:"prevHash"`
	Timestamp    int64          `json:"timestamp"`
	Nonce        int            `json:"nonce"`
	Transactions []*Transaction `json:"transactions"`
	Data         *SmartContract `json:"smartcontract"`
	BlockHash    []byte         `json:"hash"`
}

func NewBlock(prevHash []byte, nonce int, blockNumber uint64, smartContract *SmartContract) *Block {
	block := new(Block)
	block.PrevHash = prevHash
	block.Nonce = nonce
	block.BlockNumber = blockNumber
	block.Timestamp = time.Now().Unix()
	block.Data = smartContract
	block.BlockHash = block.Hash()
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

func (b *Block) Hash() []byte {
	h := sha256.New()

	// Append the timestamp.
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(b.Timestamp))
	h.Write(timestampBytes)

	// Append block number
	blockNumberBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(blockNumberBytes, uint64(b.BlockNumber))
	h.Write(blockNumberBytes)

	if b.Data != nil {
		// Append data code length as bytes
		dataCodeLengthBytes := intToBytes(len(b.Data.Code))
		h.Write(dataCodeLengthBytes)

		// Append code as bytes
		h.Write(b.Data.Code)
		// Append data as bytes
		dataDataLengthBytes := intToBytes(len(b.Data.Data))
		h.Write(dataDataLengthBytes)

		h.Write(b.Data.Data)
	}

	h.Write(intToBytes(len(b.PrevHash)))
	h.Write(b.PrevHash) // include hash for all previous methods in current hashing scheme.

	hashed := h.Sum(nil) // use bytes since it may have to have byte related operation.
	return hashed

}

func intToBytes(num int) []byte { // convert int to bytes in BigEndian
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(num))
	return bytes
}

func (b *Block) AddTransactionToTheBlock(txn *Transaction) {
	// check if the txn verification is a success or a failure
	if txn.Status == constants.TXN_VERIFICATION_SUCCESS {
		txn.Status = constants.SUCCESS
	} else {
		txn.Status = constants.FAILED
	}

	b.Transactions = append(b.Transactions, txn)
}

func (b *Block) MineBlock() {
	nonce := 0
	desiredHash := strings.Repeat("0", constants.MINING_DIFFICULTY)
	for {
		guessHash := b.Hash()
		ourSolutionHash := hex.EncodeToString(guessHash[:constants.MINING_DIFFICULTY])
		if ourSolutionHash == desiredHash {
			return
		}
		nonce++
		b.Nonce = nonce
	}

}
