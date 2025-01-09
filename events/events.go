// events/events.go
package events

import (
	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/transaction"
)

// Define event types for blockchain, consensus, peer updates, etc.
type BlockAddedEvent struct {
	Block *block.Block // Assuming you have Block defined in blockchain package
}

type TransactionAddedEvent struct {
	Transaction *transaction.Transaction
}
