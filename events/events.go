// events/events.go
package events

import (
	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/transaction"
	"fmt"
)

// Define event types for blockchain, consensus, peer updates, etc.
type BlockAddedEvent struct {
	Block *block.Block // Assuming you have Block defined in blockchain package
}

func Broadcaster(blockAdded <-chan BlockAddedEvent, transactionAdded <-chan TransactionAddedEvent) {
	for {
		select {
		case event := <-blockAdded:
			// Handle the block added event
			fmt.Println("Block added:", event.Block)
		case event := <-transactionAdded:
			// Handle the transaction added event
			fmt.Println("Transaction added:", event.Transaction)
		}
	}
}

type TransactionAddedEvent struct {
	Transaction *transaction.Transaction
	BlockHash   string
	BlockNumber uint64

	Block                  *block.Block
	Peer                   string
	Timestamp              int64
	PeerAddress            string
	PeerID                 string
	PeerPort               string
	PublicKey              string
	Signature              []byte
	From                   string
	To                     string
	Value                  uint64
	Data                   []byte
	Status                 string
	TransactionHash        string
	Nonce                  int
	PrevHash               string
	Hash                   string
	TransactionPool        []transaction.TransactionPool
	Amount                 []uint64
	Success                bool
	Failed                 bool
	Pending                string
	TxnVerification        string
	BlockchainStatus       string
	PeerBroadcastPauseTime int
	PeerPingPauseTime      int
	TxnBroadcastPauseTime  int
	FetchLastNBlocks       int
	ConsensusPauseTime     int
	newTxn                 *transaction.Transaction
}
