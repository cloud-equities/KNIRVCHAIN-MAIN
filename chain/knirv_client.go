package chain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// KNIRVClient represents a client that interacts with the KNIRVCHAIN.
type KNIRVClient struct {
	KNIRVChainAddress string           // Address of the KNIRVCHAIN API
	SmartContracts    []*SmartContract // List of smart contracts
	mutex             sync.Mutex       // Mutex for thread safety (if needed)
	rpcEndpoint       string
	chainID           string
	contractAddress   string       // Example: Ethereum contract address
	httpClient        *http.Client // For making requests to the KNIRVCHAIN
}

// NewKNIRVClient creates a new KNIRVClient.
func NewKNIRVClient(knirvchainAddress, rpcEndpoint, chainID, contractAddress string, smartContracts []*SmartContract) *KNIRVClient {
	return &KNIRVClient{
		KNIRVChainAddress: knirvchainAddress,
		SmartContracts:    smartContracts,
		rpcEndpoint:       rpcEndpoint,
		chainID:           chainID,
		contractAddress:   contractAddress,
		httpClient:        &http.Client{}, // Initialize the HTTP client
	}
}

func (c *KNIRVClient) handleGetConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		config := map[string]string{
			"REACT_APP_RPC_ENDPOINT":     c.rpcEndpoint,
			"REACT_APP_CHAIN_ID":         c.chainID,
			"REACT_APP_CONTRACT_ADDRESS": c.contractAddress,
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(config)
		if err != nil {
			log.Println("failed to encode config: ", err)
			http.Error(w, "Failed to encode config", http.StatusInternalServerError)
			return
		}
	}
}

// GetBlockchain gets blockchain from KNIRVCHAIN via API
func (kc *KNIRVClient) GetBlockchain() (*BlockchainStruct, error) {
	resp, err := http.Get(kc.KNIRVChainAddress + "/blockchain") // Example endpoint
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain: %w", err)
	}
	defer resp.Body.Close()

	var blockchain BlockchainStruct
	if err := json.NewDecoder(resp.Body).Decode(&blockchain); err != nil {
		return nil, fmt.Errorf("failed to decode blockchain: %w", err)
	}

	return &blockchain, nil
}

// startAPI starts the KNIRVClient's API server.
func (kc *KNIRVClient) StartInnerAPI(ctx context.Context) { // Add context for graceful shutdown
	r := mux.NewRouter()

	r.HandleFunc("/config", kc.GetConfig()).Methods("GET")
	r.HandleFunc("/contractListing", kc.GetContractListing()).Methods("GET")
	r.HandleFunc("/mint", kc.BroadcastMint()).Methods("POST")
	r.HandleFunc("/burn", kc.BroadcastBurn()).Methods("POST")

	server := &http.Server{
		Addr:    ":8081", // Uses a different port than the Locker Stream
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start config API server: %v", err)
		}
	}()

}
func (kc *KNIRVClient) GetConfig() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
func (kc *KNIRVClient) GetContractListing() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
func (kc *KNIRVClient) BroadcastMint() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
func (kc *KNIRVClient) BroadcastBurn() func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}

func (kc *KNIRVClient) printConnectionInfo() {
	log.Println("KNIRV Chain Info:")
	log.Println("REACT_APP_RPC_ENDPOINT=" + kc.rpcEndpoint)
	log.Println("REACT_APP_CHAIN_ID=" + kc.chainID)
	log.Println("REACT_APP_CONTRACT_ADDRESS=" + kc.contractAddress)
}

func generateChainID() string {
	hash := time.Now().UnixNano()
	return fmt.Sprintf("%x", hash)
}

func generateContractAddress() string {
	hash := time.Now().UnixNano()
	byteValue := make([]byte, 8)
	byteValue[0] = byte((hash >> 56) & 0xFF)
	byteValue[1] = byte((hash >> 48) & 0xFF)
	byteValue[2] = byte((hash >> 40) & 0xFF)
	byteValue[3] = byte((hash >> 32) & 0xFF)
	byteValue[4] = byte((hash >> 24) & 0xFF)
	byteValue[5] = byte((hash >> 16) & 0xFF)
	byteValue[6] = byte((hash >> 8) & 0xFF)
	byteValue[7] = byte((hash >> 0) & 0xFF)
	return hex.EncodeToString(byteValue)
}

// printConnectionInfo prints the locker's connection information using log.Println.
func (l *KNIRVClient) PrintConnectionInfo() { // Use uppercase first letter for exported function

	log.Println("Locker Connection Info:")
	log.Println("RPC Endpoint:", l.rpcEndpoint)
	log.Println("Chain ID:", l.chainID)
	log.Println("Contract Address:", l.contractAddress)
}

//  Methods related to smart contracts and NRNs

// MintNFT mints a new NFT via the KNIRVCHAIN.
func (lc *KNIRVClient) MintNFT(data []byte, metadata map[string]interface{}, owner string) (*NFT, error) {
	// Create the NFT locally.
	nft, err := NewNFT(data, metadata, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to create NFT: %w", err)
	}

	// Send a request to the KNIRVCHAIN to mint the NFT.
	// ... (Implementation for sending mint request to KNIRVCHAIN API) ...

	return nft, nil
}

// GetNRNByNFTID retrieves the NRN associated with an NFT via KNIRV chain
func (lc *KNIRVClient) GetNRNByNFTID(nftID string) (*NRN, error) {

	// Send a request to the KNIRVCHAIN to get the NRN by NFT ID
	// ... (Implementation to get NRN from KNIRVCHAIN)

	return nil, fmt.Errorf("GetNRNByNFTID not yet implemented")
}

// MintEmptyNRN returns an empty NRN with placeholder data
func (lc *KNIRVClient) MintEmptyNRN() (*NRN, error) {

	nrn := NewNRN("", "", big.NewInt(0), big.NewInt(0), "")
	return nrn, nil
}

// GetLatestBlock fetches the latest block from the KNIRVCHAIN.
func (lc *KNIRVClient) GetLatestBlock() (*Block, error) {
	// Send a request to the KNIRVCHAIN to get the latest block.
	// ... (Implementation for retrieving latest block from KNIRVCHAIN API) ...
	return nil, fmt.Errorf("GetLatestBlock not yet implemented") // Placeholder
}

// GetBlockData retrieves smart contract data from the given block
func (lc *KNIRVClient) GetBlockData(block *Block) (*SmartContract, error) {
	// ... implementation to extract data from Block
	return nil, nil
}

// Graceful shutdown logic
