package consensus

import (
	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/blockchain"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/events"
	"KNIRVCHAIN-MAIN/peerManager"
	"KNIRVCHAIN-MAIN/transaction"
	"log"
	"sync"
	"time"
)

// ConsensusManager manages the blockchain consensus.
type ConsensusManager struct {
	Blockchain  *blockchain.BlockchainStruct // Direct pointer to the blockchain
	PeerManager *peerManager.PeerManager     // Direct pointer to the peer manager
}

// NewConsensusManager creates a new ConsensusManager.
func NewConsensusManager(blockchain *blockchain.BlockchainStruct, peerManager *peerManager.PeerManager) *ConsensusManager {
	return &ConsensusManager{
		Blockchain:  blockchain,
		PeerManager: peerManager,
	}
}

var pm *peerManager.PeerManager
var once sync.Once

func GetPeerManager(blockAdded <-chan events.BlockAddedEvent, transactionAdded <-chan events.TransactionAddedEvent) *peerManager.PeerManager {
	once.Do(func() {
		pm = &peerManager.PeerManager{
			Peers:                        make(map[string]peerManager.Peer),
			TransactionPool:              []*transaction.Transaction{},
			Blocks:                       []*peerManager.RemoteBlock{},
			Address:                      "",
			MiningLocked:                 false,
			Mutex:                        sync.Mutex{},
			BlockNumber:                  0,
			PrevHash:                     "",
			Timestamp:                    0,
			Nonce:                        0,
			BlockAddedSubscription:       blockAdded,
			TransactionAddedSubscription: transactionAdded,
		}
	})
	return pm
}

// RunConsensus runs the blockchain consensus algorithm.
func (cm *ConsensusManager) RunConsensus() {
	for {
		if cm.Blockchain.MiningLocked {
			time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)
			continue
		}

		log.Println("Starting the consensus algorithm...")

		cm.Blockchain.MiningLocked = true // Lock before updating

		longestChain := cm.Blockchain.Blocks
		longestChainIsOurs := true

		// Fetch the last N blocks from each peer

		for peer, status := range cm.PeerManager.Peers {

			if peer != cm.PeerManager.Address && status.Status {

				remotePeerManager, err := peerManager.FetchLastNBlocks(peer)

				if err != nil {
					log.Println("Error fetching blocks:", err)
					continue
				}
				// Access remotePeerManager.Blocks
				if len(remotePeerManager.Blocks) > 0 && (len(longestChain) == 0 || pm.TransactionPool[len(pm.TransactionPool)-1].BlockNumber > longestChain[len(longestChain)-1].BlockNumber) {
					if cm.PeerManager.VerifyLastNBlocks(remotePeerManager.Blocks) {
						longestChain = make([]*block.Block, len(remotePeerManager.Blocks))
						for i, rb := range remotePeerManager.Blocks {
							block := &block.Block{
								BlockNumber:  rb.BlockNumber,
								Nonce:        rb.Nonce,
								PrevHash:     rb.PrevHash,
								Timestamp:    rb.Timestamp,
								Transactions: rb.Transactions,
							}
							longestChain[i] = block

						}

						longestChainIsOurs = false
						defer func() { cm.Blockchain.MiningLocked = false }() // Unlock after updating
					} else {
						log.Println("Chain verification failed from peer:", peer)
					}

				}

			}
		}

		if !longestChainIsOurs {

			cm.Blockchain.MiningLocked = true                     // Lock before updating
			defer func() { cm.Blockchain.MiningLocked = false }() // Unlock after updating

			newBlocks := []*block.Block{}

			for _, rb := range longestChain {
				newBlock := &block.Block{
					BlockNumber:  rb.BlockNumber,
					Nonce:        rb.Nonce,
					PrevHash:     rb.PrevHash,
					Timestamp:    rb.Timestamp,
					Transactions: rb.Transactions,
				}

				newBlocks = append(newBlocks, newBlock)
			}
			cm.Blockchain.Blocks = newBlocks            // Update the blockchain
			err := blockchain.PutIntoDb(*cm.Blockchain) // Save to DB after successful update
			if err != nil {
				log.Printf("Failed to save updated blockchain to DB: %s", err) // Log and continue, consensus will retry
			} else {
				log.Println("Updated our blockchain to the longest chain")
				for _, b := range cm.Blockchain.Blocks {
					for _, txn := range b.Transactions {
						cm.PeerManager.Broadcaster.BroadcastTransaction(txn, cm.PeerManager.Address)

					}
				}
			}

			// After updating blockchain and before sending a block added event.

		} else {
			log.Println("Our chain is the longest chain")

		}

		time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)

	}

}
