package blockchain

import (
	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/events"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"KNIRVCHAIN-MAIN/peerManager"
	"KNIRVCHAIN-MAIN/transaction"
)

type BlockchainStruct struct {
	TransactionPool  []*transaction.Transaction         `json:"transaction_pool"`
	Blocks           []*block.Block                     `json:"block_chain"`
	Address          string                             `json:"address"`
	Peers            map[string]bool                    `json:"peers"`
	MiningLocked     bool                               `json:"mining_locked"`
	Broadcaster      transaction.TransactionBroadcaster `json:"-"`
	BlockAdded       chan events.BlockAddedEvent        `json:"-"`
	TransactionAdded chan events.TransactionAddedEvent  `json:"-"`
}

var mutex sync.Mutex

func NewBlockchain(genesisBlock block.Block, address string, broadcaster transaction.TransactionBroadcaster) *BlockchainStruct {
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
		blockchainStruct.TransactionPool = []*transaction.Transaction{}
		blockchainStruct.Blocks = []*block.Block{}
		blockchainStruct.Blocks = append(blockchainStruct.Blocks, &genesisBlock)
		blockchainStruct.Address = address
		blockchainStruct.Peers = map[string]bool{}
		blockchainStruct.MiningLocked = false
		blockchainStruct.BlockAdded = make(chan events.BlockAddedEvent)
		blockchainStruct.TransactionAdded = make(chan events.TransactionAddedEvent)
		blockchainStruct.Broadcaster = broadcaster
		err := PutIntoDb(*blockchainStruct)
		if err != nil {
			panic(err.Error())
		}
		return blockchainStruct
	}
}

func NewBlockchainFromSync(remoteBlocks []*peerManager.RemoteBlock, address string, broadcaster transaction.TransactionBroadcaster) *BlockchainStruct {
	// 1. Convert RemoteBlock to Block: Deep copy is essential to avoid modification issues
	blocks := make([]*block.Block, len(remoteBlocks))
	for i, rb := range remoteBlocks {
		block := &block.Block{
			BlockNumber:  rb.BlockNumber,
			Nonce:        rb.Nonce,
			PrevHash:     rb.PrevHash,
			Timestamp:    rb.Timestamp,
			Transactions: rb.Transactions,
		}

		blocks[i] = block
	}

	bc := &BlockchainStruct{
		Blocks:       blocks,
		Address:      address, // Your blockchain node's address
		Peers:        make(map[string]bool),
		Broadcaster:  broadcaster, // Your transaction broadcaster
		MiningLocked: false,       // Add other necessary fields

	}
	return bc
}

func (bc BlockchainStruct) ToJson() string {
	nb, err := json.Marshal(bc)

	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (bc *BlockchainStruct) AddBlock(b *block.Block) {
	mutex.Lock()
	defer mutex.Unlock()

	m := map[string]bool{}
	for _, txn := range b.Transactions {
		m[txn.TransactionHash] = true
	}

	// remove txn from txn pool
	newTxnPool := []*transaction.Transaction{}
	for _, txn := range bc.TransactionPool {
		_, ok := m[txn.TransactionHash]
		if !ok {
			newTxnPool = append(newTxnPool, txn)
		}
	}

	bc.TransactionPool = newTxnPool
	bc.Blocks = append(bc.Blocks, b)

	// save the blockchain to our database
	err := PutIntoDb(*bc)
	if err != nil {
		panic(err.Error())
	}
	bc.BlockAdded <- events.BlockAddedEvent{Block: b} // Send the event to the event channel
}

func (bc *BlockchainStruct) appendTransactionToTheTransactionPool(transaction *transaction.Transaction) {
	mutex.Lock()
	defer mutex.Unlock()

	bc.TransactionPool = append(bc.TransactionPool, transaction)

	// save the blockchain to our database
	err := PutIntoDb(*bc)
	if err != nil {
		panic(err.Error())
	}
}

func (bc *BlockchainStruct) AddTransactionToTransactionPool(transaction *transaction.Transaction) {

	for _, txn := range bc.TransactionPool {
		if txn.TransactionHash == transaction.TransactionHash {
			return
		}
	}

	log.Println("Adding txn to the Transaction pool")

	valid1 := transaction.VerifyTxn()

	valid2 := bc.simulatedBalanceCheck(valid1, transaction)

	if valid1 && valid2 {
		transaction.Status = constants.TXN_VERIFICATION_SUCCESS
	} else {
		transaction.Status = constants.TXN_VERIFICATION_FAILURE
	}

	transaction.PublicKey = ""

	bc.appendTransactionToTheTransactionPool(transaction)

	bc.BroadcastLocalTransaction(transaction)

}
func (bc *BlockchainStruct) BroadcastLocalTransaction(txn *transaction.Transaction) {
	bc.Broadcaster.BroadcastTransaction(txn, bc.Address)                  // Use the interface
	bc.TransactionAdded <- events.TransactionAddedEvent{Transaction: txn} // Send the event to the event channel
}

func (bc *BlockchainStruct) simulatedBalanceCheck(valid1 bool, transaction *transaction.Transaction) bool {
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

func (bc *BlockchainStruct) ProofOfWorkMining(minersAddress string) {
	log.Println("Starting to Mine...")
	// calculate the prevHash
	nonce := 0
	for {
		if bc.MiningLocked {
			continue
		}

		prevHash := bc.Blocks[len(bc.Blocks)-1].Hash()

		if bc.MiningLocked {
			continue
		}

		// start with a nonce
		// create a new block
		guessBlock := block.NewBlock(prevHash, nonce, uint64(len(bc.Blocks)))

		if bc.MiningLocked {
			continue
		}
		// copy the transaction pool
		for _, txn := range bc.TransactionPool {

			if bc.MiningLocked {
				continue
			}

			newTxn := new(transaction.Transaction)
			newTxn.Data = txn.Data
			newTxn.From = txn.From
			newTxn.To = txn.To
			newTxn.Status = txn.Status
			newTxn.Timestamp = txn.Timestamp
			newTxn.Value = txn.Value
			newTxn.TransactionHash = txn.TransactionHash
			newTxn.PublicKey = txn.PublicKey
			newTxn.Signature = txn.Signature

			guessBlock.AddTransactionToTheBlock(newTxn)
		}

		if bc.MiningLocked {
			continue
		}

		rewardTxn := transaction.NewTransaction(constants.BLOCKCHAIN_ADDRESS, minersAddress, constants.MINING_REWARD, []byte{})
		rewardTxn.Status = constants.SUCCESS
		guessBlock.Transactions = append(guessBlock.Transactions, rewardTxn)

		if bc.MiningLocked {
			continue
		}

		// guess the Hash
		guessHash := guessBlock.Hash()
		desiredHash := strings.Repeat("0", constants.MINING_DIFFICULTY)
		ourSolutionHash := guessHash[2 : 2+constants.MINING_DIFFICULTY]

		if bc.MiningLocked {
			continue
		}

		if ourSolutionHash == desiredHash {

			if !bc.MiningLocked {
				bc.AddBlock(guessBlock)
				log.Println("Mined block number:", guessBlock.BlockNumber)
			}
			nonce = 0
			continue
		}

		nonce++
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

func (bc *BlockchainStruct) GetAllTxns() []transaction.Transaction {

	nTxns := []transaction.Transaction{}

	for i := len(bc.TransactionPool) - 1; i >= 0; i-- {
		nTxns = append(nTxns, *bc.TransactionPool[i])
	}

	txns := []transaction.Transaction{}

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
