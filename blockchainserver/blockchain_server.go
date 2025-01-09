package blockchainserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"KNIRVCHAIN-MAIN/blockchain"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/transaction"
)

type BlockchainServer struct {
	Port          uint64                       `json:"port"`
	BlockchainPtr *blockchain.BlockchainStruct `json:"blockchain"`
	Server        *http.Server
	MiningLocked  bool `json:"mining_locked"`
}

func NewBlockchainServer(port uint64, blockchainPtr *blockchain.BlockchainStruct) *BlockchainServer {
	bcs := new(BlockchainServer)
	bcs.Port = port
	bcs.BlockchainPtr = blockchainPtr
	bcs.Server = &http.Server{Addr: fmt.Sprintf(":%d", bcs.Port)}

	return bcs
}

func (bcs *BlockchainServer) GetBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		blocks := bcs.BlockchainPtr.Blocks
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(blocks); err != nil {
			http.Error(w, "Failed to marshal blocks to json", http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetBalance(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		addr := req.URL.Query().Get("address")
		x := struct {
			Balance uint64 `json:"balance"`
		}{
			bcs.BlockchainPtr.CalculateTotalCrypto(addr),
		}

		mBalance, err := json.Marshal(x)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(mBalance)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetAllNonRewardedTxns(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		txnList := bcs.BlockchainPtr.GetAllTxns()
		byteSlice, err := json.Marshal(txnList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(byteSlice)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transactions = bcs.BlockchainPtr.TransactionPool
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(transactions)
	if err != nil {
		http.Error(w, "Failed to marshal transactions to json", http.StatusInternalServerError)
		return
	}
}

func (bcs *BlockchainServer) SendTxnToTheBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodPost {
		var txn transaction.Transaction

		if err := json.NewDecoder(req.Body).Decode(&txn); err != nil {
			http.Error(w, "Invalid transaction format", http.StatusBadRequest)
			return
		}
		//Verify Transaction
		if !txn.VerifyTxn() {
			http.Error(w, "Invalid Txn Signature", http.StatusBadRequest)
			return
		}
		err := bcs.BlockchainPtr.AddTransaction(txn)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to add transaction: %v", err), http.StatusInternalServerError)
			log.Printf("Failed to add transaction: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(txn); err != nil {
			http.Error(w, "Failed to encode txn", http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
	}
}

func CheckStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		io.WriteString(w, constants.BLOCKCHAIN_STATUS)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) SendPeersList(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodPost {
		peersMap, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println("Error reading Peers")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var peersList map[string]bool
		err = json.Unmarshal(peersMap, &peersList)
		if err != nil {
			log.Println("Error Unmarshalling the Peers")

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bcs.BlockchainPtr.PeerManager.UpdatePeers(peersList)
		res := map[string]string{}
		res["status"] = "success"
		x, err := json.Marshal(res)
		if err != nil {
			log.Println("Error while marshalling the http Response to be sent.")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(x)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) FetchLastNBlocks(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		blocks := bcs.BlockchainPtr.Blocks
		blockchain1 := new(blockchain.BlockchainStruct)
		if len(blocks) < constants.FETCH_LAST_N_BLOCKS {
			blockchain1.Blocks = blocks
		} else {
			blockchain1.Blocks = blocks[len(blocks)-constants.FETCH_LAST_N_BLOCKS:]
		}
		blockJSON, err := json.Marshal(blockchain1.Blocks)
		if err != nil {
			http.Error(w, "Failed to marshal blocks to json", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(blockJSON)

	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Start() {
	http.HandleFunc("/", bcs.GetBlockchain)
	http.HandleFunc("/balance", bcs.GetBalance)
	http.HandleFunc("/blocks", bcs.GetBlockchain)
	http.HandleFunc("/get_all_non_rewarded_txns", bcs.GetAllNonRewardedTxns)
	http.HandleFunc("/send_txn", bcs.SendTxnToTheBlockchain)
	http.HandleFunc("/transactions", bcs.handleGetTransactions)
	http.HandleFunc("/send_peers_list", bcs.SendPeersList)
	http.HandleFunc("/check_status", CheckStatus)
	http.HandleFunc("/fetch_last_n_blocks", bcs.FetchLastNBlocks)
	log.Println("Launching webserver at port :", bcs.Port)
	go func() {
		if err := bcs.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Failed to start blockchain server: %v", err)
		}
	}()
}

func (bcs *BlockchainServer) Stop() {
	if err := bcs.Server.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

func init() {
	log.SetPrefix(constants.BLOCKCHAIN_NAME + ":")
}
