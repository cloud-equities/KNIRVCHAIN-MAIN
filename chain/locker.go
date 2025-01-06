package chain

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

// Locker represents a P2P streaming locker.  Includes KNIRVClient.
type Locker struct {
	ChainPath       string
	ID              string
	OwnerID         string
	NFTID           string
	DB              *bolt.DB
	Stream          *Stream
	Blockchain      *BlockchainStruct
	rpcEndpoint     string
	chainID         string
	contractAddress string
	kc              *KNIRVClient // KNIRVClient field
}

// NewLocker creates a new Locker. Pass necessary configuration parameters including KNIRVClient.
func NewLocker(id, ownerID, nftID, dbPath, host string, port int, blockchain *BlockchainStruct, rpcEndpoint, chainID, contractAddress, knirvchainAddress string, smartContracts []*SmartContract) (*Locker, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open boltdb: %w", err)
	}
	stream, err := NewStream(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize p2p stream: %w", err)
	}

	kc := NewKNIRVClient(knirvchainAddress, rpcEndpoint, chainID, contractAddress, smartContracts)

	locker := &Locker{
		ID:              id,
		OwnerID:         ownerID,
		NFTID:           nftID,
		DB:              db,
		Stream:          stream,
		Blockchain:      blockchain,
		rpcEndpoint:     rpcEndpoint,
		chainID:         chainID,
		contractAddress: contractAddress,
		kc:              kc, // Initialize the KNIRVClient
	}

	return locker, nil
}

// GetBlockchain returns the SmartContract associated with the locker.
func (l *Locker) GetOwnerID() *SmartContract {
	// Assuming you are storing the SmartContract in the locker
	// and initializing it during Locker creation.
	// Otherwise, you'll need a way to set the blockchain.
	return l.GetOwnerID()
}

// startAPI starts the locker's API server.
func (l *Locker) startOuterAPI() {
	r := mux.NewRouter() // You might want to pass the router from outside if you are already using it.

	r.HandleFunc("/config", l.GetConfig()).Methods("GET")
	r.HandleFunc("/nrn", l.GetNRN()).Methods("GET")
	r.HandleFunc("/keychain", l.KeyChain()).Methods("POST")

	go func() {
		server := &http.Server{
			Addr:    ":8084", // Uses a different port than the KNIRVClient
			Handler: r,
		}

		err := server.ListenAndServe()
		if err != nil {
			// Use a better logging mechanism than log.Fatal here,
			// or handle the error gracefully.
			log.Fatal("Failed to start config API server: ", err)
		}

	}()

}

func (l *Locker) KeyChain() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
func (l *Locker) GetConfig() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
func (l *Locker) GetNRN() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}

// StartStreaming starts the P2P streaming server
func (l *Locker) StartStreaming(data []byte) error {
	return l.Stream.Start(data)
}

// StopStreaming stops the P2P streaming server
func (l *Locker) StopStreaming() error {
	return l.Stream.Stop()
}

// GetLatestVersion retrieves the latest version of the data
func (l *Locker) GetLatestVersion() ([]byte, error) {
	var data []byte
	err := l.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("versions"))
		if b == nil {
			return fmt.Errorf("versions bucket not found")
		}
		latest := b.Get([]byte("latest"))
		if latest == nil {
			return fmt.Errorf("latest version not found")
		}

		data = b.Get(latest)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version from db: %w", err)
	}
	return data, nil
}

// AddKeyChain adds a new KeyChain to the locker's database
func (l *Locker) AddKeyChain(keyChainID string) error {
	err := l.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("padlocks"))
		if err != nil {
			return fmt.Errorf("failed to create padlock bucket: %w", err)
		}
		err = b.Put([]byte(keyChainID), []byte(time.Now().String()))
		if err != nil {
			return fmt.Errorf("failed to add padlock to db: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update db: %w", err)
	}
	return nil
}

// RecordView records a view of the locker with the padlock id, and the viewer id
func (l *Locker) RecordView(keyChainID string, viewerID string) error {
	err := l.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("views"))
		if err != nil {
			return fmt.Errorf("failed to create views bucket: %w", err)
		}

		key := fmt.Sprintf("%s-%s", keyChainID, viewerID)
		err = b.Put([]byte(key), []byte(time.Now().String()))
		if err != nil {
			return fmt.Errorf("failed to record view in db: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}
	return nil
}

// AddVersion adds a new version of the data to the Locker database, with the NRN contract address
func (l *Locker) AddVersion(data []byte, contractAddress string) error {
	err := l.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("versions"))
		if err != nil {
			return fmt.Errorf("failed to create versions bucket: %w", err)
		}

		key := fmt.Sprintf("%x", time.Now().UnixNano())
		err = b.Put([]byte(key), data)
		if err != nil {
			return fmt.Errorf("failed to save new version: %w", err)
		}

		err = b.Put([]byte("latest"), []byte(key))
		if err != nil {
			return fmt.Errorf("failed to save latest version: %w", err)
		}

		nrnBucket, err := tx.CreateBucketIfNotExists([]byte("nrns"))
		if err != nil {
			return fmt.Errorf("failed to create nrns bucket: %w", err)
		}
		err = nrnBucket.Put([]byte("latest"), []byte(contractAddress))
		if err != nil {
			return fmt.Errorf("failed to save latest nrn: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add version to db: %w", err)
	}
	return nil
}
