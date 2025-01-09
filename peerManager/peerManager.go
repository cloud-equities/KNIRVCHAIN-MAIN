package peerManager

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/events"
	"KNIRVCHAIN-MAIN/transaction"
)

type PeerTransactionBroadcaster struct {
	// You might need fields here to access the peer list, etc.  Possibly a reference to the PeerManager itself.
	PeerManager *PeerManager
}

type PeerManager struct {
	Peers                            map[string]Peer `json:"peers"`
	Peer                             Peer            `json:"peer"`
	TransactionPool                  []*Transaction  `json:"transaction_pool"`
	Blocks                           []*RemoteBlock  `json:"block_chain"`
	Address                          string          `json:"address"`
	MiningLocked                     bool            `json:"mining_locked"`
	Mutex                            sync.Mutex      `json:"mutex"`
	BlockNumber                      uint64          `json:"block_number"`
	PrevHash                         string          `json:"prevHash"`
	Timestamp                        int64           `json:"timestamp"`
	Nonce                            int             `json:"nonce"`
	Transactions                     []*Transaction  `json:"transactions"`
	BlockAddedSubscription           <-chan events.BlockAddedEvent
	TransactionAddedSubscription     <-chan events.TransactionAddedEvent
	BlockAddedSubscriptionChan       chan events.BlockAddedEvent
	TransactionAddedSubscriptionChan chan events.TransactionAddedEvent
	Broadcaster                      PeerTransactionBroadcaster
	BlockAdded                       chan events.BlockAddedEvent
	TransactionAdded                 chan events.TransactionAddedEvent
	PeerTransactionBroadcaster       PeerTransactionBroadcaster
}
type Peer struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Status    bool   `json:"status"`
	LastPing  int64  `json:"last_ping"`
	LastTx    int64  `json:"last_tx"`
	LastBlock int64  `json:"last_block"`
	LastCons  int64  `json:"last_cons"`
	LastSync  int64  `json:"last_sync"`
	LastFetch int64  `json:"last_fetch"`
}

type Peers struct {
	LastPeer      int64           `json:"last_peer"`
	LastTxSync    int64           `json:"last_tx_sync"`
	LastBlockSync int64           `json:"last_block_sync"`
	LastConsSync  int64           `json:"last_cons_sync"`
	LastFetchSync int64           `json:"last_fetch_sync"`
	LastPeerSync  int64           `json:"last_peer_sync"`
	PeersList     map[string]bool `json:"peersList"`
}

type ClientPeer struct {
	PeerAddress string `json:"address"`
	ID          string `json:"id"`
}

var pm *PeerManager
var once sync.Once

func GetPeerManager(blockAdded <-chan events.BlockAddedEvent, transactionAdded <-chan events.TransactionAddedEvent) *PeerManager {
	once.Do(func() {
		pm = &PeerManager{
			Peers:                        make(map[string]Peer),
			TransactionPool:              []*Transaction{},
			Blocks:                       []*RemoteBlock{},
			Address:                      "",
			MiningLocked:                 false,
			Mutex:                        sync.Mutex{},
			BlockNumber:                  0,
			PrevHash:                     "",
			Timestamp:                    0,
			Nonce:                        0,
			Transactions:                 []*Transaction{},
			BlockAddedSubscription:       blockAdded,
			TransactionAddedSubscription: transactionAdded,
		}
	})
	return pm
}

type RemoteBlockchainStruct struct {
	TransactionPool []*transaction.TransactionPool `json:"transaction_pool"`
	Blocks          []*RemoteBlock                 `json:"block_chain"`
	Address         string                         `json:"address"`
	Peers           map[string]bool                `json:"peers"`
	MiningLocked    bool                           `json:"mining_locked"`
}

type Transaction struct {
	From            string                        `json:"from"`
	To              string                        `json:"to"`
	Value           uint64                        `json:"value"`
	Data            []byte                        `json:"data"`
	Status          string                        `json:"status"`
	Timestamp       int64                         `json:"timestamp"`
	TransactionHash string                        `json:"transaction_hash"`
	PublicKey       string                        `json:"public_key,omitempty"`
	Signature       []byte                        `json:"Signature"`
	TransactionPool []transaction.TransactionPool `json:"transaction_pool"`
}

type RemoteBlock struct {
	BlockNumber  uint64                     `json:"block_number"`
	PrevHash     string                     `json:"prevHash"`
	Timestamp    int64                      `json:"timestamp"`
	Nonce        int                        `json:"nonce"`
	Transactions []*transaction.Transaction `json:"transactions"`
}

func (pm *PeerManager) StartListening() { // New function to listen for events
	go func() {
		for {
			select {
			case event := <-pm.BlockAddedSubscription:
				// Handle the BlockAdded event. NO mutex needed, blockchain handled it.
				pm.processBlockAdded(event.Block)

			case event := <-pm.TransactionAddedSubscription:
				// Handle the TransactionAddedEvent
				pm.processTransactionAdded(event.Transaction)

				// ... cases for other events
			}
		}
	}()
}

func (pm *PeerManager) processBlockAdded(block *block.Block) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	pm.Blocks = append(pm.Blocks, &RemoteBlock{
		BlockNumber:  block.BlockNumber,
		PrevHash:     block.PrevHash,
		Timestamp:    block.Timestamp,
		Nonce:        block.Nonce,
		Transactions: block.Transactions,
	})
	err := PutIntoDb(*pm)
	if err != nil {
		panic(err.Error())
	}
}
func (pm *PeerManager) processTransactionAdded(transaction *Transaction) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	pm.TransactionPool = append(pm.TransactionPool, &Transaction{
		TransactionHash: transaction.TransactionHash,
		From:            transaction.From,
		To:              transaction.To,
		Data:            transaction.Data,
		Timestamp:       transaction.Timestamp,
		Status:          transaction.Status,
	})
	err := PutIntoDb(*pm)
	if err != nil {
		panic(err.Error())
	}
}
func (pm *PeerManager) convertBlockToRemoteBlock(block *block.Block) *RemoteBlock {
	// ... conversion logic (similar to what you had before)
	return &RemoteBlock{ /* ... */ }
}

func (pm *PeerManager) GetBlockchain() []*RemoteBlock {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	return pm.Blocks
}

func (pm *PeerManager) GetTransactionPool() []*Transaction {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	return pm.TransactionPool
}
func (pm *PeerManager) GetBlockchainLength() int {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	return len(pm.Blocks)
}
func (pm *PeerManager) GetTransactionPoolLength() int {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	return len(pm.TransactionPool)
}

//	func (pm *PeerManager) UpdateBlockchain(remoteBlocks []*RemoteBlock) {
//		pm.Mutex.Lock()
//		defer pm.Mutex.Unlock()
//		for _, remoteBlock := range remoteBlocks {
//			block := Block{
//				BlockNumber:  remoteBlock.BlockNumber,
//				PrevHash:     remoteBlock.PrevHash,
//				Timestamp:    remoteBlock.Timestamp,
//				Nonce:        remoteBlock.Nonce,
//				Transactions: remoteBlock.Transactions,
//			}
//			pm.Blocks = append(pm.Blocks, &block)
//		}
//		err := PutIntoDb(*pm)
//		if err != nil {
//			panic(err.Error())
//		}
//	}
func (pm *PeerManager) UpdateTransactionPool(Transactions []*Transaction) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	for _, Transaction := range Transactions {
		Transaction := &Transaction{
			TransactionHash: Transaction.TransactionHash,
			From:            Transaction.From,
			To:              Transaction.To,
			Data:            Transaction.Data,
			Timestamp:       Transaction.Timestamp,
			Status:          Transaction.Status,
			PublicKey:       Transaction.PublicKey,
			Signature:       Transaction.Signature,
		}
		pm.TransactionPool = append(pm.TransactionPool, Transaction)
	}
	err := PutIntoDb(*pm)
	if err != nil {
		panic(err.Error())
	}
}
func (pm *PeerManager) UpdatePeers(peersList map[string]bool) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	for peerID, status := range peersList {
		if peer, exists := pm.Peers[peerID]; exists {
			peer.Status = status
			pm.Peers[peerID] = peer
		} else {
			pm.Peers[peerID] = Peer{
				ID:      peerID,
				Address: "", // You might need to set the address appropriately
				Status:  status,
			}
			log.Println("Updating Peers List..", peersList)

		}
	}
	err := PutIntoDb(*pm)
	if err != nil {
		panic(err.Error())
	}

}
func (pm *PeerManager) UpdatePeer(peer Peer) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	pm.Peer = peer
	err := PutIntoDb(*pm)
	if err != nil {
		panic(err.Error())
	}

}

func SyncBlockchain(address string) (*PeerManager, error) {
	log.Println("Started syncing blockchain from node:", address)
	ourURL := fmt.Sprintf("%s/", address)
	resp, err := http.Get(ourURL)
	if err != nil {
		return nil, err
	}

	// read the body of the response here
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pms PeerManager
	err = json.Unmarshal(data, &pms)
	if err != nil {
		return nil, err
	}

	log.Println("Finished syncing blockchain from node:", address)

	return &pms, nil
}

func (pm *PeerManager) SendPeersList(address string) {
	data := pm.PeersToJson()
	ourURL := fmt.Sprintf("%s/send_peers_list", address)
	http.Post(ourURL, "application/json", bytes.NewBuffer(data))
}

func (rbc Transaction) BlockToJson() string {
	nb, err := json.Marshal(rbc)

	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (pm *PeerManager) PeersToJson() []byte {
	nb, _ := json.Marshal(pm.Peers)
	return nb
}

func (pm *PeerManager) CheckStatus(address string) bool {
	ourURL := fmt.Sprintf("%s/check_status", address)
	resp, err := http.Get(ourURL)
	if err != nil {
		log.Println(err)
		return false
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()

	return string(data) == constants.BLOCKCHAIN_STATUS
}

func (pm *PeerManager) BroadcastPeerList() {
	for peer, status := range pm.Peers {
		if peer != pm.Address && status.Status {
			pm.SendPeersList(peer)
			time.Sleep(constants.PEER_BROADCAST_PAUSE_TIME * time.Second)
		}
	}
}

func (pm *PeerManager) DialAndUpdatePeers() {
	for {
		log.Println("Pinging Peers", pm.Peers)
		newList := pm.Peers

		for peer := range newList {
			if peer != pm.Address {
				peerStatus := pm.CheckStatus(peer)
				peer := newList[peer]
				peer.Status = peerStatus
				newList[peer.Address] = peer
			} else {
				newList[peer] = Peer{
					ID:      peer,
					Address: pm.Address,
					Status:  true,
				}
			}
		}

		// update our peers List
		peerStatusMap := make(map[string]bool)
		for peer, status := range newList {
			peerStatusMap[peer] = status.Status
		}
		pm.UpdatePeers(peerStatusMap)
		log.Println("Updated Peer status : ", pm.Peers)

		// broadcast our new peers list
		pm.BroadcastPeerList()

		time.Sleep(constants.PEER_PING_PAUSE_TIME * time.Second)
	}
}

// For LocalTransaction

func (pm *PeerManager) SendTxnToThePeer(address string, txn *Transaction) {
	data := txn.BlockToJson()
	ourURL := fmt.Sprintf("%s/send_txn", address)
	http.Post(ourURL, "application/json", strings.NewReader(data))
}

func (pm *PeerTransactionBroadcaster) BroadcastTransaction(txn *Transaction, excludeAddress string) {
	for peer, status := range pm.PeerManager.Peers { // Access peer list somehow
		if peer != excludeAddress && status.Status {
			// ... logic for sending the transaction (like your current SendTxnToThePeer)
			log.Println("Broadcasting LocalTransaction to the peer:", peer, "Transaction:", txn.BlockToJson())

			pm.PeerManager.SendTxnToThePeer(peer, txn)
			time.Sleep(constants.TXN_BROADCAST_PAUSE_TIME * time.Second)

		}
	}
}

func FetchLastNBlocks(address string) (*PeerManager, error) {
	log.Println("Fetching last", constants.FETCH_LAST_N_BLOCKS, "blocks")
	ourURL := fmt.Sprintf("%s/fetch_last_n_blocks", address)
	resp, err := http.Get(ourURL)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var nbc PeerManager
	err = json.Unmarshal(data, &nbc)
	if err != nil {
		return nil, err
	}

	return &nbc, nil
}

func (rb RemoteBlock) RemoteHash() string {

	bs, _ := json.Marshal(rb)
	sum := sha256.Sum256(bs)
	hexRep := hex.EncodeToString(sum[:32])
	formattedHexRep := constants.HEX_PREFIX + hexRep

	return formattedHexRep
}
func verifyLastNBlocks(chain []*RemoteBlock) bool {
	if chain[0].BlockNumber != 0 && chain[0].RemoteHash()[2:2+constants.MINING_DIFFICULTY] != strings.Repeat("0", constants.MINING_DIFFICULTY) {
		log.Println("Chain verification failed for block", chain[0].BlockNumber, "hash", chain[0].RemoteHash())
		return false
	}

	for i := 1; i < len(chain); i++ {
		if chain[i-1].RemoteHash() != chain[i].PrevHash {
			log.Println("Failed to verify prevHash for block number", chain[i].BlockNumber)
			return false
		}

		if chain[i].RemoteHash()[2:2+constants.MINING_DIFFICULTY] != strings.Repeat("0", constants.MINING_DIFFICULTY) {
			log.Println("Chain verification failed for block", chain[0].BlockNumber, "hash", chain[0].RemoteHash())
			return false
		}
	}

	return true
}

func (pm *PeerManager) UpdateBlockchain(chain []*RemoteBlock) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	blocks := []*RemoteBlock{}
	initBlockNumber := chain[0].BlockNumber + 1
	if len(pm.Blocks) > int(initBlockNumber) {
		for _, block := range pm.Blocks[:initBlockNumber] {
			blocks = append(blocks, block)
		}
	}
	log.Println("Updating our blockchain from block number", initBlockNumber)
	for _, block := range pm.Blocks[:initBlockNumber] {
		blocks = append(blocks, &RemoteBlock{*&pm.BlockNumber, *&pm.PrevHash, *&pm.Timestamp, *&pm.Nonce, pm.Transactions}, block)
	}
	blocks = append(blocks, chain...)

	pm.Blocks = make([]*RemoteBlock, len(blocks))
	for i, block := range blocks {
		pm.Blocks[i] = block
	}

	// update the transaction pool
	found := map[string]bool{}
	for _, txn := range pm.TransactionPool {
		found[txn.Transaction.TransactionHash] = false
	}

	for _, block := range chain {
		for _, txn := range block.Transactions {
			_, ok := found[txn.Transaction.TransactionHash]
			if ok {
				found[txn.Transaction.TransactionHash] = true
			}
		}
	}

	newTxnPool := []*Transaction{}
	for _, txn := range pm.TransactionPool {
		if !found[txn.Transaction.TransactionHash] {
			newTxnPool = append(newTxnPool, txn)
		}
	}

	pm.TransactionPool = make([]*Transaction, len(newTxnPool))
	for i, txn := range newTxnPool {
		pm.TransactionPool[i] = txn
	}

	// save the blockchain in the database
	err := PutIntoDb(*pm)
	if err != nil {
		panic(err.Error())
	}
}
func (pm *PeerManager) FetchLastNBlocksFromPeer(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		peer := req.URL.Query().Get("peer")
		bc1, err := FetchLastNBlocks(peer)
		if err != nil {
			log.Println("Error while  fetching last n blocks from peer:", peer, "Error:", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, ToJson())
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (pm *PeerManager) AddPeer(p Peer) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	_, ok := pm.Peers[p.ID]
	if ok {
		return
	}

	pm.Peers[p.ID] = p

}

func (pm *PeerManager) RemovePeer(id string) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	delete(pm.Peers, id)
}

func (pm *PeerManager) GetPeers() map[string]Peer {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	return pm.Peers
}

func (pm *PeerManager) GetPeer(id string) (Peer, bool) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	peer, ok := pm.Peers[id]
	return peer, ok
}

func (pm *PeerManager) String() string {
	bytes, _ := json.Marshal(pm.Peers)
	return string(bytes)
}
