package transactionBroadcaster

type BlockAddedEvent struct {
	// Define fields here
}

type TransactionAddedEvent struct {
	TransactionEvent BlockAddedEvent
}

type TransactionBroadcaster interface {
	BroadcastTransaction(TransactionAddedEvent)
}
