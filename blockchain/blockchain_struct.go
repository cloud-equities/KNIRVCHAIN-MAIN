package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"KNIRVCHAIN-MAIN/constants"

	"KNIRVCHAIN-MAIN/peerManager"

	"KNIRVCHAIN-MAIN/transactionBroadcaster"
)

type BlockchainStruct struct {
	TransactionPool  []*Transaction                                    `json:"transaction_pool"`
	Blocks           []*Block                                          `json:"block_chain"`
	Address          string                                            `json:"address"`
	Peers            map[string]bool                                   `json:"peers"`
	MiningLocked     bool                                              `json:"mining_locked"`
	Broadcaster      transactionBroadcaster.TransactionBroadcaster     `json:"-"`
	BlockAdded       chan transactionBroadcaster.BlockAddedEvent       `json:"-"`
	TransactionAdded chan transactionBroadcaster.TransactionAddedEvent `json:"-"`
	PeerManager      *peerManager.PeerManager                          `json:"-"`
	Mutex            sync.Mutex                                        `json:"-"`
}

var mutex sync.Mutex

func NewBlockchain(genesisBlock Block, address string, broadcaster transactionBroadcaster.TransactionBroadcaster, peerManager *peerManager.PeerManager) *BlockchainStruct {
	exists, _ := KeyExists()

	if exists {

		blockchainStruct, err := GetBlockchain()
		blockchainStruct.Broadcaster = broadcaster
		if err != nil {
			panic(err.Error())
		}
		return blockchainStruct
	} else {
		blockchainStruct := new(BlockchainStruct)
		blockchainStruct.TransactionPool = []*Transaction{}
		blockchainStruct.Blocks = []*Block{}
		blockchainStruct.Blocks = append(blockchainStruct.Blocks, &genesisBlock)
		blockchainStruct.Address = address
		blockchainStruct.Peers = map[string]bool{}
		blockchainStruct.MiningLocked = false
		blockchainStruct.BlockAdded = make(chan transactionBroadcaster.BlockAddedEvent)
		blockchainStruct.TransactionAdded = make(chan transactionBroadcaster.TransactionAddedEvent)
		blockchainStruct.Broadcaster = broadcaster
		blockchainStruct.PeerManager = peerManager
		blockchainStruct.Mutex = sync.Mutex{}
		//	err := PutIntoDb(*blockchainStruct)
		//if err != nil {
		//	panic(err.Error())
		//}
		return blockchainStruct
	}
}

func NewBlockchainFromSync(remoteBlocks []*peerManager.RemoteBlock, address string, broadcaster transactionBroadcaster.TransactionBroadcaster, peerManager *peerManager.PeerManager) *BlockchainStruct {
	// 1. Convert RemoteBlock to Block: Deep copy is essential to avoid modification issues
	blocks := make([]*Block, len(remoteBlocks))
	for i, rb := range remoteBlocks {
		block := &Block{
			BlockNumber: rb.BlockNumber,
			Nonce:       rb.Nonce,
			PrevHash:    rb.PrevHash,
			Timestamp:   rb.Timestamp,
		}

		blocks[i] = block
	}

	bc := &BlockchainStruct{
		Blocks:       blocks,
		Address:      address, // Your blockchain node's address
		Peers:        make(map[string]bool),
		Broadcaster:  broadcaster, // Your transaction broadcaster
		MiningLocked: false,       // Add other necessary fields
		PeerManager:  peerManager,
		Mutex:        sync.Mutex{},
	}
	return bc
}

func (bc *BlockchainStruct) ToJson() string {
	nb, err := json.Marshal(bc)

	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (bc *BlockchainStruct) AddBlock(b *Block) {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	m := map[string]bool{}
	for _, txn := range b.Transactions {
		m[txn.Hash()] = true
	}

	// remove txn from txn pool
	newTxnPool := []*Transaction{}
	for _, txn := range bc.TransactionPool {
		_, ok := m[txn.Hash()]
		if !ok {
			newTxnPool = append(newTxnPool, txn)
		}
	}

	bc.TransactionPool = newTxnPool

	bc.Blocks = append(bc.Blocks, b)

	// save the blockchain to our database
	err := PutIntoDb(bc)
	if err != nil {
		panic(err.Error())
	}
	//bc.BlockAdded <- events.BlockAddedEvent{Block: transaction} // Send the event to the event channel
	log.Printf("Block added: %+v", b) // Log only if necessary
}

func (bc *BlockchainStruct) appendTransactionToTheTransactionPool(transaction *Transaction) {
	mutex.Lock()
	defer mutex.Unlock()

	bc.TransactionPool = append(bc.TransactionPool, transaction)

	// save the blockchain to our database
	err := PutIntoDb(bc)
	if err != nil {
		panic(err.Error())
	}
}

func (bc *BlockchainStruct) AddTransactionToTransactionPool(transaction *Transaction) {

	for _, txn := range bc.TransactionPool {
		if txn.Hash() == transaction.Hash() {
			return
		}
	}

	log.Println("Adding txn to the Transaction pool")

	valid1 := transaction.VerifyTxn()

	//valid2 := bc.simulatedBalanceCheck(valid1, transaction)

	if valid1 {
		transaction.Status = constants.TXN_VERIFICATION_SUCCESS
	} else {
		transaction.Status = constants.TXN_VERIFICATION_FAILURE
	}

	transaction.PublicKey = ""

	bc.appendTransactionToTheTransactionPool(transaction)

	bc.BroadcastLocalTransaction(transaction)

}
func (bc *BlockchainStruct) BroadcastLocalTransaction(txn *Transaction) {
	//bc.Broadcaster.BroadcastTransaction(txn, bc.Address)

	// Log only if needed:
	log.Println("Broadcasting Transaction", txn) // Use the interface
	//bc.TransactionAdded <- events.TransactionAddedEvent{Transaction: txn} // Send the event to the event channel
}

func (bc *BlockchainStruct) simulatedBalanceCheck(valid1 bool, transaction *Transaction) bool {
	balance := bc.CalculateTotalCrypto(transaction.From)
	for _, txn := range bc.TransactionPool {
		if transaction.From == txn.From && valid1 {
			if balance >= txn.Value {
				balance -= txn.Value
			} else {
				break
			}
		}
	}

	return balance >= transaction.Value
}

func (bc *BlockchainStruct) MineNewBlock(minersAddress string) (*Block, error) {
	bc.MiningLocked = true // Lock mining during block creation

	defer func() { bc.MiningLocked = false }() // Unlock *always*, even if error

	prevHash := bc.Blocks[len(bc.Blocks)-1].Hash()
	newBlock := NewBlock(prevHash, 0, uint64(len(bc.Blocks))) // nonce starts at 0

	// Deep copy transactions from the pool
	for _, txn := range bc.TransactionPool {
		newTxn := NewTransaction(txn.From, txn.To, txn.Value, txn.Data)
		newTxn.Timestamp = txn.Timestamp
		newTxn.Status = txn.Status
		newTxn.Signature = txn.Signature
		newTxn.PublicKey = txn.PublicKey // Ensure public key is copied
		if err := newBlock.AddTransactionToTheBlock(newTxn); err != nil {
			return nil, fmt.Errorf("failed to add transaction to block: %w", err)
		}

	}

	rewardTxn := NewTransaction(constants.BLOCKCHAIN_ADDRESS, minersAddress, constants.MINING_REWARD, []byte{})
	rewardTxn.Status = constants.SUCCESS
	if err := newBlock.AddTransactionToTheBlock(rewardTxn); err != nil {
		return nil, fmt.Errorf("failed to add reward transaction: %w", err)
	}

	if err := newBlock.Mine(constants.MINING_DIFFICULTY); err != nil {
		return nil, fmt.Errorf("mining error: %w", err)
	}
	return newBlock, nil
}

func (bc *BlockchainStruct) ProofOfWorkMining(minersAddress string, stopMining <-chan bool, miningStopped chan<- bool) {
	for {

		if bc.MiningLocked {
			time.Sleep(constants.MINING_PAUSE_TIME * time.Second)
			continue
		}

		select {
		case <-stopMining:

			miningStopped <- true
			return

		default:

			newBlock, err := bc.MineNewBlock(minersAddress)
			if err != nil {
				log.Println("Error Mining Block: ", err)

				continue // Don't crash, try again in the next iteration
			}

			if !bc.MiningLocked {
				bc.AddBlock(newBlock)
				log.Println("Mined block number:", newBlock.BlockNumber)
				bc.BroadcastLocalTransaction(newBlock.Transactions[len(newBlock.Transactions)-1])

			}

		}

	}

}

func (bc *BlockchainStruct) CalculateTotalCrypto(address string) uint64 {
	sum := uint64(0)

	for _, blocks := range bc.Blocks {
		for _, txns := range blocks.Transactions {
			if txns.Status == constants.SUCCESS {
				if txns.To == address {
					sum += txns.Value
				} else if txns.From == address {
					sum -= txns.Value
				}
			}
		}
	}
	return sum
}

func (bc *BlockchainStruct) GetAllTxns() []Transaction {

	nTxns := []Transaction{}

	for i := len(bc.TransactionPool) - 1; i >= 0; i-- {
		nTxns = append(nTxns, *bc.TransactionPool[i])
	}

	txns := []Transaction{}

	for _, blocks := range bc.Blocks {
		for _, txn := range blocks.Transactions {
			if txn.From != constants.BLOCKCHAIN_ADDRESS {
				txns = append(txns, *txn)
			}
		}
	}
	for i := len(txns) - 1; i >= 0; i-- {
		nTxns = append(nTxns, txns[i])
	}

	return nTxns
}
func (bc *BlockchainStruct) AddTransaction(txn Transaction) error {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()
	if bc.MiningLocked {
		return fmt.Errorf("mining is locked, cannot add transaction")
	}

	if !txn.VerifyTxn() {
		return fmt.Errorf("txn verification failed")
	}

	bc.TransactionPool = append(bc.TransactionPool, &txn)
	//bc.TransactionAdded <- events.TransactionAddedEvent{Transaction: &txn} // Send event *after* adding to pool
	return nil

}
