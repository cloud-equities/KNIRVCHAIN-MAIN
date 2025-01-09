// events/events.go
package events

import (
	"KNIRVCHAIN-MAIN/blockchain"
)

// Define event types for blockchain, consensus, peer updates, etc.
type BlockAddedEvent struct {
	Block *blockchain.Block // Assuming you have Block defined in blockchain package
}

type TransactionAddedEvent struct {
	Transaction *blockchain.Transaction
}
