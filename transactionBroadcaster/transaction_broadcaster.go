package transactionBroadcaster

type TransactionBroadcaster interface {
	BroadcastTransaction(txn *Transaction, excludeAddress string)
}
