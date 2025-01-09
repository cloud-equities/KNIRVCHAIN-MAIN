package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/blockchain"
	"KNIRVCHAIN-MAIN/consensus"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/events"
	"KNIRVCHAIN-MAIN/peerManager"
	"KNIRVCHAIN-MAIN/transaction"
	"KNIRVCHAIN-MAIN/utils"
	"KNIRVCHAIN-MAIN/walletserver"
)

func init() {
	log.SetPrefix(constants.BLOCKCHAIN_NAME + ":")
}

// MockBlockchainServer to inject blockchain data
type MockBlockchainServer struct {
	BlockchainPtr *blockchain.BlockchainStruct
}

func NewMockBlockchainServer(blockchain *blockchain.BlockchainStruct) *MockBlockchainServer {
	return &MockBlockchainServer{BlockchainPtr: blockchain}
}
func (bcs *MockBlockchainServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {
	case "/blocks":
		if r.Method == http.MethodGet {
			bcs.handleGetBlocks(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/transactions":
		if r.Method == http.MethodGet {
			bcs.handleGetTransactions(w, r)
		} else if r.Method == http.MethodPost {
			bcs.handlePostTransaction(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/peers":
		if r.Method == http.MethodPost {
			bcs.handlePostPeer(w, r)
		} else if r.Method == http.MethodGet {
			bcs.handleGetPeers(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func (bcs *MockBlockchainServer) handlePostPeer(w http.ResponseWriter, r *http.Request) {
	var peer peerManager.Peer
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&peer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bcs.BlockchainPtr.Peers[peer.Address] = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(peer)
}

func (bcs *MockBlockchainServer) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bcs.BlockchainPtr.Peers)
}

func (bcs *MockBlockchainServer) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	blocks := bcs.BlockchainPtr.Blocks
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(blocks); err != nil {
		http.Error(w, "Failed to marshal blocks to json", http.StatusInternalServerError)
		return
	}

}

func (bcs *MockBlockchainServer) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transactions = bcs.BlockchainPtr.TransactionPool
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(transactions)
	if err != nil {
		http.Error(w, "Failed to marshal transactions to json", http.StatusInternalServerError)
		return
	}
}

func (bcs *MockBlockchainServer) handlePostTransaction(w http.ResponseWriter, r *http.Request) {
	var t transaction.Transaction

	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bcs.BlockchainPtr.TransactionAdded <- events.TransactionAddedEvent{Transaction: &t}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// TestMain serves as entrypoint for the tests
func TestMain(m *testing.M) {
	//Load .env.test file
	err := godotenv.Load("./test.env")
	if err != nil {
		log.Fatalf("Error loading .env.test file: %v", err)
	}
	os.Exit(m.Run()) //Run tests after loading environment variables
}

// Helper function to set up the test environment
func setupTestEnv(t *testing.T) (*Config, *httptest.Server, *consensus.ConsensusManager, *peerManager.PeerManager, chan bool, chan bool, chan bool, chan bool) {
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Port = 5001 //For testing purposes force the port
	// Set up mock blockchain and server
	genesisBlock := block.NewBlock("0x0", 0, 0)
	blockAddedChan := make(chan events.BlockAddedEvent)
	transactionAddedChan := make(chan events.TransactionAddedEvent)
	pm := consensus.GetPeerManager(blockAddedChan, transactionAddedChan)
	pm.Address = "http://127.0.0.1:" + strconv.Itoa(int(cfg.Port))
	pm.Broadcaster = peerManager.PeerTransactionBroadcaster{PeerManager: pm}
	bc := blockchain.NewBlockchain(*genesisBlock, pm.Address, &pm.Broadcaster, pm)
	bc.Peers[bc.Address] = true
	mockServer := httptest.NewServer(NewMockBlockchainServer(bc))

	consensusMgr := consensus.NewConsensusManager(bc, pm)

	startMining := make(chan bool)
	startConsensus := make(chan bool)
	stopMining := make(chan bool)
	miningStopped := make(chan bool)

	return cfg, mockServer, consensusMgr, pm, startMining, startConsensus, stopMining, miningStopped
}

// Helper function to tear down the test environment
func tearDownTestEnv(server *httptest.Server) {
	server.Close()
}

func Test_ChainSubcommand(t *testing.T) {
	// Set up test environment
	cfg, _, _, _, _, _, _, _ := setupTestEnv(t)

	// Start test node as a separate process
	cmd, err := StartTestNode(int(cfg.Port), "0x123", "")
	if err != nil {
		t.Fatalf("Failed to start test node: %v", err)
	}
	defer func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait() // wait for process to terminate
		}
	}()

	time.Sleep(7 * time.Second) // Give time to start and mine blocks

	// Fetch Blocks
	url := fmt.Sprintf("http://127.0.0.1:%d/blocks", cfg.Port)

	var blocks []*block.Block
	// Retry logic to handle eventual consistency
	for i := 0; i < 5; i++ { // Retry up to 5 times
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to get blocks: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK { // Only unmarshal if status is OK
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			err = json.Unmarshal(body, &blocks)

			if err == nil {
				break // Exit retry loop if unmarshal successful
			} else {
				t.Logf("Attempt %d: Failed to unmarshal blocks: %v. Retrying...", i+1, err)
				time.Sleep(time.Second) // Wait before retrying
			}
		} else {

			t.Logf("Attempt %d: Expected status OK but got %v. Retrying...", i+1, resp.StatusCode)
			time.Sleep(time.Second)

		}

	}
	if len(blocks) == 0 {
		t.Fatalf("Test failed, unable to retrieve blocks")
	}
	// Rest of test logic
	if len(blocks) <= 1 {
		t.Error("Expected more blocks to be mined, but there isn't")
	}

}

func Test_WalletSubcommand(t *testing.T) {
	// Set up test environment
	cfg, mockServer, _, _, _, _, _, _ := setupTestEnv(t)
	defer tearDownTestEnv(mockServer)

	// Start Wallet Server
	walletPort := 8081
	blockchainNodeAddress := fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
	ws := walletserver.NewWalletServer(uint64(walletPort), blockchainNodeAddress)
	stop := ws.Start()
	t.Cleanup(func() {
		stop()
	})

	time.Sleep(time.Second)

	// Test: Create a transaction via wallet
	url := fmt.Sprintf("http://127.0.0.1:%d/send_signed_txn", walletPort) // Send to send_signed_txn, not /transactions
	jsonData := map[string]interface{}{
		"from_address": "0x234",
		"to_address":   "0x567",
		"value":        10,
	}
	jsonValue, _ := json.Marshal(jsonData)

	// Add the private key to the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
	}
	//q := req.URL.Query()
	//q.Add("privateKey", "YOUR_PRIVATE_KEY_HERE") // Replace with an actual test private key!
	//req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to post a transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created but got: %v", resp.StatusCode)
	}

	// Check if transaction went to the transaction pool
	url = fmt.Sprintf("http://127.0.0.1:%d/transactions", cfg.Port)

	var transactions []*transaction.Transaction
	// Retry logic to handle eventual consistency
	for i := 0; i < 5; i++ {

		resp, err = http.Get(url)
		if err != nil {
			t.Fatalf("Failed to fetch transactions: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			t.Logf("Attempt %d: Status code: %v, body: %s", i+1, resp.StatusCode, body)

			err = json.Unmarshal(body, &transactions)
			if err == nil {
				if len(transactions) >= 1 { //Check that at least one transaction is present.
					break
				} else {
					t.Logf("Attempt %d: No transactions in the pool. Retrying...", i+1)

					time.Sleep(time.Second)
				}

			} else {
				t.Logf("Attempt %d: Failed to unmarshal transactions: %v. Retrying...", i+1, err)
				time.Sleep(time.Second)

			}

		} else {
			t.Logf("Attempt %d: Expected status OK but got %v, body: %s. Retrying...", i+1, resp.StatusCode, string(body))
			time.Sleep(time.Second)
		}

	}
	if len(transactions) == 0 {
		t.Fatalf("Test failed, unable to retrieve transactions")
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected one transaction but got %v", len(transactions))
	}
}

func Test_PeerSync(t *testing.T) {
	// Set up test environment
	cfg, _, _, _, _, _, _, _ := setupTestEnv(t)

	//Start Node 1 (Use StartTestNode as before)
	cmd1, err := StartTestNode(int(cfg.Port), "0x123", "")
	if err != nil {
		t.Fatalf("Failed to start test node 1: %v", err)
	}
	defer func() {
		if cmd1 != nil && cmd1.Process != nil {
			cmd1.Process.Kill()
			cmd1.Wait()
		}
	}()

	//Start Node 2
	ts2Port := int(cfg.Port) + 2 // Define port for the second node (Use a different port like 5002)
	ts2, err := StartTestNode(ts2Port, "0x456", fmt.Sprintf("http://127.0.0.1:%d", cfg.Port))
	if err != nil {
		t.Fatalf("Failed to start test node 2: %v", err)
		if ts2 != nil && ts2.Process != nil {
			ts2.Process.Kill()
			ts2.Wait()
		}
		return
	}
	defer func() {
		if ts2 != nil && ts2.Process != nil {
			ts2.Process.Kill()
			ts2.Wait()
		}
	}()

	time.Sleep(10 * time.Second)

	// Fetch blocks from Node 1
	url1 := fmt.Sprintf("http://127.0.0.1:%d/blocks", cfg.Port)

	var blocks1 []*block.Block
	// Retry logic for node 1
	for i := 0; i < 5; i++ {
		resp1, err := http.Get(url1)
		if err != nil {
			t.Fatalf("Failed to fetch blocks from node 1: %v", err)
		}
		defer resp1.Body.Close()
		if resp1.StatusCode == http.StatusOK {
			body1, err := io.ReadAll(resp1.Body)
			if err != nil {
				t.Fatalf("Failed to read response body from node 1 %v", err)
			}
			err = json.Unmarshal(body1, &blocks1)
			if err == nil {
				break
			} else {
				t.Logf("Attempt %d: Failed to unmarshal blocks from node 1: %v. Retrying...", i+1, err)
				time.Sleep(time.Second)

			}
		} else {
			t.Logf("Attempt %d: Expected status OK from node 1 but got: %v. Retrying...", i+1, resp1.StatusCode)
			time.Sleep(time.Second)
		}
	}
	if len(blocks1) == 0 {
		t.Fatalf("Test failed, unable to retrieve blocks from node 1")
	}

	// Fetch blocks from Node 2
	url2 := fmt.Sprintf("http://127.0.0.1:%d/blocks", ts2Port)
	var blocks2 []*block.Block
	// Retry logic for node 2
	for i := 0; i < 5; i++ {
		resp2, err := http.Get(url2)
		if err != nil {
			t.Fatalf("Failed to fetch blocks from node 2: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode == http.StatusOK {
			body2, err := io.ReadAll(resp2.Body)
			if err != nil {
				t.Fatalf("Failed to read response body from node 2 %v", err)
			}
			err = json.Unmarshal(body2, &blocks2)
			if err == nil {
				break
			} else {
				t.Logf("Attempt %d: Failed to unmarshal blocks from node 2: %v. Retrying...", i+1, err)
				time.Sleep(time.Second)
			}
		} else {
			t.Logf("Attempt %d: Expected status OK from node 2 but got: %v. Retrying...", i+1, resp2.StatusCode)
			time.Sleep(time.Second)
		}
	}
	if len(blocks2) == 0 {
		t.Fatalf("Test failed, unable to retrieve blocks from node 2")
	}

	if len(blocks1) != len(blocks2) {
		t.Fatalf("Expected the nodes to be sync but got node 1 blocks: %v and node 2 blocks: %v", len(blocks1), len(blocks2))
	}
	if !utils.CompareBlocks(blocks1, blocks2) {
		t.Fatalf("Expected the blocks to be the same but they are not")
	}
}
