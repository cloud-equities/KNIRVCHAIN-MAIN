package chain

import (
	"fmt"
	"sync"
	"time"

	"github.com/cloud-equities/KNIRVCHAIN/log"
)

type ConsensusManager struct {
	mutex            sync.Mutex
	longestChainHash string
	MiningLocked     bool
	updateRequired   bool
	pauseConsensus   chan bool
	sync             chan bool
}

// Creates a new consensus manager object
func NewConsensusManager() *ConsensusManager {
	return &ConsensusManager{
		MiningLocked:   false,
		updateRequired: false,
		pauseConsensus: make(chan bool),
		sync:           make(chan bool),
	}
}

// RunConsensus implements the consensus function
func (cm *ConsensusManager) RunConsensus(bc *BlockchainStruct) {
	fmt.Errorf("Running consensus...")
	for {
		cm.mutex.Lock()
		if !cm.updateRequired { // Only do the update, if a blockchain sync has not requested to sync.
			ourHash, err := bc.getOurCurrentBlockHash()
			if err != nil {
				log.LogError("Unable to resolve our latest block hash", err)
				cm.mutex.Unlock()                          // We have not successfully ran a chain compare
				time.Sleep(time.Duration(5 * time.Second)) // Add back off for when we fail
				continue
			}

			if ourHash != cm.longestChainHash { // if the blockchain has changed then use the new change
				fmt.Errorf("My chain is not the longest")

				valid := true // do real checks before using another chain.

				if valid {
					bc.updateBlockchain(cm)
				}
			} else { // If the hash matches and it has not changed
				fmt.Errorf("My chain is longest, thus I am not updating my blockchain") // do nothing, just keep our version.

			}
		}
		cm.mutex.Unlock() // ensure the resource is released
		time.Sleep(time.Duration(10 * time.Second))
	}
}

func (cm *ConsensusManager) lockMining() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.MiningLocked = true
}

func (cm *ConsensusManager) unlockMining() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.MiningLocked = false
}
func (cm *ConsensusManager) setUpdateRequired(val bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.updateRequired = val
}
func (cm *ConsensusManager) getMiningLockState() bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.MiningLocked
}

func (cm *ConsensusManager) getSyncState() chan bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.sync
}

func (cm *ConsensusManager) setLongestChain(chainHash string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.longestChainHash = chainHash
}

func (cm *ConsensusManager) getPauseSignal() chan bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.pauseConsensus
}
