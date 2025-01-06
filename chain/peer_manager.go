package chain

import (
	"encoding/json"
	"fmt"

	"github.com/cloud-equities/KNIRVCHAIN/constants"

	"log"
	"net/http"
	"sync"
	"time"
)

type Peer struct {
	ID          string `json:"id"`
	PeerAddress string `json:"address"`
}

type PeerManager struct {
	Peers map[string]Peer `json:"peers"`
	mutex sync.Mutex
}

var pm *PeerManager
var once sync.Once

// SyncBlockchain synchronizes the blockchain from a remote node
func SyncBlockchain(remoteNodeAddress string, db *LevelDB, address string) (*BlockchainStruct, error) {

	// Fetch the blockchain from the remote node.
	resp, err := http.Get(remoteNodeAddress + "/") // Correct address.
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blockchain from remote node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote node returned status %d", resp.StatusCode)
	}

	var remoteBlockchain BlockchainStruct

	if err := json.NewDecoder(resp.Body).Decode(&remoteBlockchain); err != nil {
		return nil, fmt.Errorf("failed to decode remote blockchain: %w", err)
	}

	// Save the fetched blockchain data to the new blockchain's database (IMPORTANT).
	err = db.PutIntoDb(&remoteBlockchain, address) // Save blockchain at 'address'
	if err != nil {
		return nil, fmt.Errorf("failed to save blockchain to db: %w", err)
	}

	return &remoteBlockchain, nil
}

func GetPeerManager() *PeerManager {
	once.Do(func() {
		pm = &PeerManager{
			Peers: map[string]Peer{},
		}
	})
	return pm
}

func (pm *PeerManager) AddPeer(p Peer) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	_, ok := pm.Peers[p.ID]
	if ok {
		return
	}

	pm.Peers[p.ID] = p

}

func (pm *PeerManager) RemovePeer(id string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	delete(pm.Peers, id)
}

func (pm *PeerManager) GetPeers() map[string]Peer {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.Peers
}

func (pm *PeerManager) GetPeer(id string) (Peer, bool) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	peer, ok := pm.Peers[id]
	return peer, ok
}

func (pm *PeerManager) String() string {
	bytes, _ := json.Marshal(pm.Peers)
	return string(bytes)
}

func (bc *BlockchainStruct) BroadcastTransaction(transaction *Transaction) {
	// get all the peers
	pm := GetPeerManager()
	peers := pm.GetPeers()

	// loop through them all, send the transaction to each of them
	for _, peer := range peers {
		go func(p Peer) {
			// do not broadcast to yourself

			if bc.ChainAddress == p.PeerAddress {
				return
			}
			log.Println("Broadcasting to peer:", p.PeerAddress)
			// create a new client

			client := NewPeerClient(p.PeerAddress)

			// send the transaction to each client
			_, err := client.BroadcastTransaction(transaction)
			if err != nil {
				fmt.Println("Failed to broadcast transaction:", err)
			}

		}(peer)
		time.Sleep(time.Millisecond * 200)
	}
}

//func (bc *BlockchainStruct) BroadcastBlock(b *Block) {
// get all the peers
//	pm := GetPeerManager()
//	peers := pm.GetPeers()

// loop through them all, send the transaction to each of them
//	for _, peer := range peers {
//		go func(p Peer) {
// do not broadcast to yourself
//			if bc.ChainAddress == p.PeerAddress {
//				return
//			}
//			log.Println("Broadcasting to peer:", p.PeerAddress)
// create a new client
//			client := NewPeerClient(p.PeerAddress)

// send the transaction to each client
//			_, err := client.BroadcastBlock(b)
//			if err != nil {
//				fmt.Println("Failed to broadcast block:", err)
//			}
//		}(peer)
//		time.Sleep(time.Millisecond * 200)
//	}
//}

//func (bc *BlockchainStruct) AddPeer(peer string) {
//	if bc.Peers == nil {
//		bc.Peers = map[string]bool{}
//	}
//	bc.Peers[peer] = true
//}

//pm := GetPeerManager()
//peer := Peer{
//	ID:          "",
//	PeerAddress: peerAddress,
//	}

//pm.AddPeer(peer)

//	log.Println("Added peer:", peerAddress)

//func (bc *BlockchainStruct) AddPeer(db *LevelDB, peerAddress string) { // Add db parameter
// ...
//	err := bc.PutIntoDb(db, bc.ChainAddress) // Pass db and chainAddress
// ...
//	log.Println("Added peer:", peerAddress)
//	pm.AddPeer(peer)
//}

//}

func (bc *BlockchainStruct) RunConsensus() {

	for {
		log.Println("Starting the consensus algorithm...")
		longestChain := bc.Blocks
		lengthOfTheLongestChain := bc.Blocks[len(bc.Blocks)-1].BlockNumber + 1
		longestChainIsOur := true
		for peer, status := range bc.Peers {
			if peer != bc.ChainAddress && status {
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

		if verifyLastNBlocks(longestChain) {
			// stop the Mining until updation
			bc.MiningLocked = true
			bc.UpdateBlockchain(longestChain)
			// restart the Mining as updation is complete
			bc.MiningLocked = false
			log.Println("Updation of Blockchain complete !!!")
		} else {
			log.Println("Chain Verification Failed, Hence not updating my blockchain")
		}

		time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)
	}

}

func (bc *BlockchainStruct) UpdateBlockchain(longestChain []*Block) {
	panic("unimplemented")
}

func FetchLastNBlocks(peer string) (*BlockchainStruct, error) {
	ourURL := fmt.Sprintf("%s/", peer)
	resp, err := http.Get(ourURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//unmarshal the response
	var bs BlockchainStruct
	return &bs, json.NewDecoder(resp.Body).Decode(&bs)
}
func verifyLastNBlocks(_ []*Block) bool {
	// Implement your verification logic here
	return true // Replace with actual verification
}
