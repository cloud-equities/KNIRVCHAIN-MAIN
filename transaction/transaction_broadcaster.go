package transaction

type TransactionBroadcaster interface {
	BroadcastTransaction(txn *Transaction, excludeAddress string)
}
