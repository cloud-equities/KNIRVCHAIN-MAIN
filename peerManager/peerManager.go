package peerManager

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/transaction"
)

type PeerTransactionBroadcaster struct {
	// You might need fields here to access the peer list, etc.  Possibly a reference to the PeerManager itself.
	PeerManager *PeerManager
}

type PeerManager struct {
	Peers           map[string]Peer `json:"peers"`
	Peer            Peer            `json:"peer"`
	TransactionPool []*Transaction  `json:"transaction_pool"`
	Blocks          []*RemoteBlock  `json:"block_chain"`
	Address         string          `json:"address"`
	MiningLocked    bool            `json:"mining_locked"`
	Mutex           sync.Mutex      `json:"mutex"`
	BlockNumber     uint64          `json:"block_number"`
	PrevHash        string          `json:"prevHash"`
	Timestamp       int64           `json:"timestamp"`
	Nonce           int             `json:"nonce"`
	Transactions    []*Transaction  `json:"transactions"`
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

func GetPeerManager() *PeerManager {
	once.Do(func() {
		pm = &PeerManager{
			Peers:           make(map[string]Peer),
			TransactionPool: []*Transaction{},
			Blocks:          []*RemoteBlock{},
			Address:         "",
			MiningLocked:    false,
			Mutex:           sync.Mutex{},
			BlockNumber:     0,
			PrevHash:        "",
			Timestamp:       0,
			Nonce:           0,
			Transactions:    []*Transaction{},
		}
	})
	return pm
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
func (pm *PeerManager) UpdateTransactionPool(Transactions []*transaction.Transaction) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	for _, Transaction := range Transactions {
		transaction := transaction.Transaction{
			TransactionHash: Transaction.TransactionHash,
			Sender:          Transaction.Sender,
			Receiver:        Transaction.Receiver,
			Amount:          Transaction.Amount,
			Timestamp:       Transaction.Timestamp,
		}
		pm.TransactionPool = append(pm.TransactionPool, &transaction)
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

type RemoteBlockchainStruct struct {
	TransactionPool []*transaction.TransactionPool `json:"transaction_pool"`
	Blocks          []*RemoteBlock                 `json:"block_chain"`
	Address         string                         `json:"address"`
	Peers           map[string]bool                `json:"peers"`
	MiningLocked    bool                           `json:"mining_locked"`
}

type Transaction struct {
	transaction.Transaction
}

type RemoteBlock struct {
	BlockNumber  uint64         `json:"block_number"`
	PrevHash     string         `json:"prevHash"`
	Timestamp    int64          `json:"timestamp"`
	Nonce        int            `json:"nonce"`
	Transactions []*Transaction `json:"transactions"`
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

func (rbc *RemoteBlockchainStruct) RunConsensus() {

	for {
		log.Println("Starting the consensus algorithm...")
		longestChain := rbc.Blocks
		lengthOfTheLongestChain := rbc.Blocks[len(rbc.Blocks)-1].BlockNumber + 1
		longestChainIsOur := true
		for peer, status := range rbc.Peers {
			if peer != rbc.Address && status {
				bc1, err := FetchLastNBlocks(peer)
				if err != nil {
					log.Println("Error while  fetching last n blocks from peer:", peer, "Error:", err.Error())
					continue
				}

				lengthOfTheFetchedChain := bc1.Blocks[len(bc1.Blocks)-1].BlockNumber + 1
				if lengthOfTheFetchedChain > lengthOfTheLongestChain {
					longestChain = bc1.Blocks
					lengthOfTheLongestChain = lengthOfTheFetchedChain
					longestChainIsOur = false
				}
			}
		}

		if longestChainIsOur {
			log.Println("My chain is longest, thus I am not updating my blockchain")
			time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)
			continue
		}

		remoteBlocks := make([]*RemoteBlock, len(longestChain))
		for i, block := range longestChain {
			remoteBlocks[i] = &RemoteBlock{
				BlockNumber:  block.BlockNumber,
				PrevHash:     block.PrevHash,
				Timestamp:    block.Timestamp,
				Nonce:        block.Nonce,
				Transactions: block.Transactions,
			}
		}
		if verifyLastNBlocks(remoteBlocks) {
			// stop the Mining until updation
			rbc.MiningLocked = true
			remoteBlocks := make([]*RemoteBlock, len(longestChain))
			for i, block := range longestChain {
				remoteBlocks[i] = &RemoteBlock{
					BlockNumber:  block.BlockNumber,
					PrevHash:     block.PrevHash,
					Timestamp:    block.Timestamp,
					Nonce:        block.Nonce,
					Transactions: block.Transactions,
				}
			}
			pm.UpdateBlockchain(remoteBlocks)
			// restart the Mining as updation is complete
			rbc.MiningLocked = false
			log.Println("Updation of Blockchain complete !!!")
		} else {
			log.Println("Chain Verification Failed, Hence not updating my blockchain")
		}

		time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)
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
