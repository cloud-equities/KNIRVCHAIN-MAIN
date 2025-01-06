package chain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloud-equities/KNIRVCHAIN/constants"
	errors "github.com/cloud-equities/KNIRVCHAIN/errors"
	knirvlog "github.com/cloud-equities/KNIRVCHAIN/log"
)

type BlockchainStruct struct {
	TransactionPool []*Transaction  `json:"transaction_pool"`
	Blocks          []*Block        `json:"block_chain"`
	ChainAddress    string          `json:"chain_address"`
	Peers           map[string]bool `json:"peers"`
	MiningLocked    bool            `json:"mining_locked"`
	OwnerAddress    string          `json:"owner_address"`
	WalletAddress   string          `json:"wallet_address"`
}

type BlockchainOptions struct {
	TransactionPool []*Transaction  `json:"transaction_pool"`
	Blocks          []*Block        `json:"block_chain"`
	ChainAddress    string          `json:"chain_address"`
	Peers           map[string]bool `json:"peers"`
	MiningLocked    bool            `json:"mining_locked"`
	OwnerAddress    string          `json:"owner_address"`
	WalletAddress   string          `json:"wallet_address"`
}

var mutex sync.Mutex

func NewBlockchain(genesisBlock *Block, chainAddress string, db *LevelDB) *BlockchainStruct {
	bc, err := CreateNewBlockchain(genesisBlock, chainAddress, db)
	if err != nil {
		knirvlog.FatalError("Failed to create chain:", err)
	}
	return bc
}
func NewLevelDB() *LevelDB {
	db, err := NewDBClient(constants.BLOCKCHAIN_DB_PATH)
	if err != nil {
		knirvlog.FatalError("Error loading Level DB data from:", err)
	}
	return db
}

func CreateNewBlockchain(genesisBlock *Block, chainAddress string, db *LevelDB) (*BlockchainStruct, error) {
	// Use the passed db.
	exists, err := db.KeyExists(chainAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to determine if key exists for: %s : %w", chainAddress, err)
	}

	if exists {
		blockchainData, err := db.GetBlockchain(chainAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get data from database for address: %s: %w", chainAddress, err)
		}
		blockchain, ok := blockchainData.(*BlockchainStruct)
		if !ok {
			return nil, fmt.Errorf("invalid blockchain data in database for chainAddress %s, struct: %v", chainAddress, blockchainData)
		}
		return blockchain, nil
	} else {
		blockchainStruct := new(BlockchainStruct)
		blockchainStruct.TransactionPool = []*Transaction{}
		blockchainStruct.Blocks = []*Block{}
		blockchainStruct.Blocks = append(blockchainStruct.Blocks, genesisBlock)
		blockchainStruct.ChainAddress = chainAddress
		blockchainStruct.Peers = map[string]bool{}
		blockchainStruct.MiningLocked = false

		err := blockchainStruct.PutIntoDb(db, chainAddress)
		if err != nil {
			return nil, fmt.Errorf("unable to put blockchain to DB: %w", err)
		}
		return blockchainStruct, nil

	}
}
func NewBlockchainFromSync(bc1 *BlockchainStruct, chainAddress string, db *LevelDB, opts ...BlockchainOptions) (*BlockchainStruct, error) { // Added options
	options := BlockchainOptions{} //Default empty options
	if len(opts) > 0 {
		options = opts[0]
	}
	optionsMiningLocked := false
	if options.MiningLocked {
		optionsMiningLocked = true
	}

	bc1.Blocks = options.Blocks
	bc1.TransactionPool = options.TransactionPool
	bc1.ChainAddress = options.ChainAddress
	bc1.Peers = options.Peers
	bc1.MiningLocked = optionsMiningLocked
	bc1.OwnerAddress = options.OwnerAddress
	bc1.WalletAddress = options.WalletAddress

	bc2 := *bc1 // get by value and dereference into local
	bc2.ChainAddress = chainAddress

	err := bc2.PutIntoDb(db, chainAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to update the blockchain object with chain ID: %s, %w", chainAddress, err) // return the newly formatted error instead.
	}
	return &bc2, nil
}

func (bc BlockchainStruct) PeersToJson() []byte {
	nb, _ := json.Marshal(bc.Peers)
	return nb
}
func (bc BlockchainStruct) ToJson() string {
	nb, err := json.Marshal(bc)
	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (bc *BlockchainStruct) AddBlock(db *LevelDB, b *Block) { // Removed *LevelDB here
	mutex.Lock()
	defer mutex.Unlock()

	m := map[string]bool{}
	for _, txn := range b.Transactions {
		m[txn.TransactionHash] = true
	}

	// remove txn from txn pool
	newTxnPool := []*Transaction{}
	for _, txn := range bc.TransactionPool {
		_, ok := m[txn.TransactionHash]
		if !ok {
			newTxnPool = append(newTxnPool, txn)
		}
	}

	bc.TransactionPool = newTxnPool
	bc.Blocks = append(bc.Blocks, b)
	// save the blockchain to our database
	err := bc.PutIntoDb(db, bc.ChainAddress)
	if err != nil {
		knirvlog.FatalError("Failed to save blockchain to database:", err)
	}
}
func (bc *BlockchainStruct) appendTransactionToTheTransactionPool(transaction *Transaction) {
	mutex.Lock()
	defer mutex.Unlock()

	bc.TransactionPool = append(bc.TransactionPool, transaction)

	// save the blockchain to our database
	err := bc.PutIntoDb(NewLevelDB(), bc.ChainAddress)
	if err != nil {
		knirvlog.FatalError("Failed to save blockchain to database:", err)
	}
}

func (bc *BlockchainStruct) AddTransactionToTransactionPool(transaction *Transaction) {

	for _, txn := range bc.TransactionPool {
		if txn.TransactionHash == transaction.TransactionHash {
			return
		}
	}

	knirvlog.LogInfo("Adding txn to the Transaction pool")
	newTxn := new(Transaction)
	newTxn.From = transaction.From
	newTxn.To = transaction.To
	newTxn.Value = transaction.Value
	newTxn.Data = transaction.Data
	newTxn.Status = transaction.Status
	newTxn.Timestamp = transaction.Timestamp
	newTxn.TransactionHash = transaction.TransactionHash
	newTxn.PublicKey = transaction.PublicKey
	newTxn.Signature = transaction.Signature

	valid1 := transaction.VerifyTxn()

	valid2 := bc.simulatedBalanceCheck(valid1, transaction)
	if valid1 && valid2 {
		transaction.Status = constants.TXN_VERIFICATION_SUCCESS
	} else {
		transaction.Status = constants.TXN_VERIFICATION_FAILURE
	}
	transaction.PublicKey = ""

	bc.appendTransactionToTheTransactionPool(transaction)

	bc.BroadcastTransaction(newTxn)

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

func (bc *BlockchainStruct) ProofOfWorkMining(minersAddress string) {
	knirvlog.LogInfo("Starting to Mine...")
	nonce := 0
	cons := NewConsensusManager()
	go func() {
		cons.RunConsensus(bc) // Implement the consensus method using a goroutine.
	}()

	for {
		if cons.getMiningLockState() { // check state and re-eval if not safe.
			time.Sleep(time.Duration(5 * time.Second))
			continue
		}

		smartContract := &SmartContract{
			Code: []byte("some smart contract code"),
			Data: []byte("some data"),
		}
		guessBlock := NewBlock([]byte{}, nonce, uint64(len(bc.Blocks)), smartContract)

		if cons.getMiningLockState() {
			time.Sleep(time.Duration(5 * time.Second)) // use non blocking checks and returns before doing more calculation intensive calls.
			continue
		}

		for _, txn := range bc.TransactionPool { // loop through list
			if cons.getMiningLockState() { //check if Mining lock state allows execution.
				time.Sleep(time.Duration(5 * time.Second))
				continue // Skip implementation because consensus state will indicate stop for data synchronization purposes
			}
			newTxn := new(Transaction) // make deep copy.
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
		if cons.getMiningLockState() {
			time.Sleep(time.Duration(5 * time.Second))
			continue
		}
		rewardTxn := NewTransaction(constants.BLOCKCHAIN_ADDRESS, minersAddress, constants.MINING_REWARD, []byte{})
		rewardTxn.Status = constants.SUCCESS
		guessBlock.Transactions = append(guessBlock.Transactions, rewardTxn)

		if cons.getMiningLockState() {
			time.Sleep(time.Duration(5 * time.Second)) // ensure we respect any calls that ask for resources to be locked when required, or when system has detected invalid data is currently present and might require cleaning up for synchronization and proper handling.
			continue                                   // check lock and sleep to skip resource hog if is currently set as `true`.
		}
		// guess the Hash
		guessHash := guessBlock.Hash()
		desiredHash := strings.Repeat("0", constants.MINING_DIFFICULTY)
		ourSolutionHash := hex.EncodeToString(guessHash[:constants.MINING_DIFFICULTY])
		if cons.getMiningLockState() {
			time.Sleep(time.Duration(5 * time.Second))
			continue
		}

		if ourSolutionHash == desiredHash {
			if !cons.getMiningLockState() {
				bc.AddBlock(NewLevelDB(), guessBlock) // create and implement new data objects or create with new structs/classes with interfaces and type constraints that you are required to utilize for future features or security concerns.
				bc.BroadcastBlock(guessBlock)
				knirvlog.LogInfo(fmt.Sprintf("Mined block number: %d", guessBlock.BlockNumber))

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

func (bc *BlockchainStruct) PutIntoDb(db *LevelDB, chainAddress string) error {
	return db.PutIntoDb(bc, chainAddress)
}

func (bc *BlockchainStruct) GetLastBlock() (*Block, error) { // Get the data for chain syncing.
	if len(bc.Blocks) == 0 {
		err := errors.New("genesis block could not be located")                // Create error object.
		knirvlog.LogError("chain is empty, genesis block not available:", err) // Log the error.
		return nil, err
	}

	return bc.Blocks[len(bc.Blocks)-1], nil // use methods where they should be for object creation and object logic manipulation to avoid bugs due to out of context calls and method signatures and naming collision errors.

}

func (bc *BlockchainStruct) DialAndUpdatePeers() { // Dialling the network using Peer Addresses, for proper sync.
	knirvlog.LogInfo("Dialling and updating peers...")
	pm := GetPeerManager() // use newly available chain and Peer components.
	peers := pm.GetPeers()
	for _, peer := range peers {
		if bc.ChainAddress == peer.PeerAddress {
			continue
		}
		bc.AddPeer(peer.PeerAddress) // you can access data by calling these values, with properly created objects, and their types.
	}
}

func (bc *BlockchainStruct) BroadcastBlock(b *Block) {
	knirvlog.LogInfo("Broadcasting newly added block to all peers!") // you were also using standard system library logging when this new library, that you had specifically implemented is available now.

	//Implement a function here that will be able to get other data based on the newly set Blockchain type which you implemented for handling chain peer, sync, or block management of type KNIRVCHAIN.
}

func (bc *BlockchainStruct) AddPeer(peer string) { // add a method using a specific interface type for new additions to other objects, where these variables have to make these references and use proper values when creating new methods in a new workflow implementation such as what you were previously implementing.

	if bc.Peers == nil { // if a type exists but that has not been yet created use the helper to create it or retrieve from a known source of your applications initial code generation steps before being assigned to this component or application structure object definition to improve type safeness of code at design and compilation stages of software development, making it safer by not trying to generate variables if they already have a correct state or an expected outcome for type definition.
		bc.Peers = map[string]bool{}
	}
	bc.Peers[peer] = true // add this peer information to the `Peers` object, now all objects can have their correct states when operating in concurrent contexts.

}
func (bc *BlockchainStruct) getOurCurrentBlockHash() (string, error) {
	lastBlock, err := bc.GetLastBlock() // call the other structs methods here as this struct only need to check, the last data set in that struct, while relying on the `GetLastBlock()` as implemented by the `BlockchainStruct`
	if err != nil {                     // Check for failure and respond to code for expected behavior in errors state
		return "", fmt.Errorf("Unable to get last block hash: %w", err)
	}
	if lastBlock == nil {
		return "", fmt.Errorf("Unable to get last block hash due to no block state available, from: %v", lastBlock)
	}
	hash := lastBlock.Hash() // use proper code from objects methods as much as is possible to avoid issues like type and parameter confusion in test environments, to promote more portable, testable, reliable code.
	return hex.EncodeToString(hash), nil

}

func (bc *BlockchainStruct) updateBlockchain(cm *ConsensusManager) { // the logic implemented in the last update created this new and useful implementation of this method using a sync method to create predictable actions during chain management.
	knirvlog.LogInfo("updating block chain, with latest block, locking mining....") // correct use of logging function for testing workflow data from implementation with project requirements for implementation of a component, struct, or variable of any declared data-type in project code structure.
	syncChan := cm.getSyncState()                                                   // use existing state variables for thread control using data channels to make program work reliably and not based on some non validated default values as parameters to create resources and manage concurrent functionality requirements in project's logic.

	cm.setUpdateRequired(true) // update value before attempting to lock mining to prevent unexpected results during implementation
	cm.lockMining()            // Lock the mining during chain update
	syncChan <- true           //Send Signal that we are performing sync, to prevent issues during chain syncing or database interactions, or methods modifying that state at an unexpected or undetermined time frame using resource blocking with data locks on struct components that require them.
	lastBlock, err := bc.GetLastBlock()
	if err != nil {
		knirvlog.LogError("unable to get last block of blockchain for peer sync:", err) //use logger and ensure correct method call.
		syncChan <- false                                                               // inform consensus object that resource was not acquired for block management or chain sync and cannot complete method or method call chains.
		return
	}

	if lastBlock == nil {
		knirvlog.LogError("unable to access last block: ", errors.New("genesis block could not be located..."))
		syncChan <- false
		return
	}

	cm.setLongestChain(hex.EncodeToString(lastBlock.Hash()))
	err = bc.PutIntoDb(NewLevelDB(), bc.ChainAddress) // make call using object to ensure methods and variables and other test settings follow requirements from new workflow implementation parameters instead of default or hardcoded parameters in implementation for code clarity and reducing runtime errors by making requirements for code to work clear with a validated state before continuing further with application testing.
	if err != nil {
		knirvlog.LogError("unable to update blockchain: ", err) // if not able to use puttod then fail gracefully
		syncChan <- false
		return // implement safety, using early returns.
	}

	time.Sleep(time.Second * 10)

	knirvlog.LogInfo("Updating Blockchain... Done!") // inform user we have stopped.
	cm.setUpdateRequired(false)                      // unlock logic, once complete
	syncChan <- false
}
