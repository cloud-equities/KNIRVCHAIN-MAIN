package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cloud-equities/KNIRVCHAIN/constants"
	"github.com/cloud-equities/KNIRVCHAIN-chain"
	lockbox "github.com/cloud-equities/KNIRVCHAIN/lockbox"
)

type Vault struct {
	Port                  uint64 `json:"port"`
	BlockchainNodeAddress string `json:"blockchain_node_addres"`
}

func (vs *Vault) CreateNewLockBox(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		lockbox1, err := lockbox.NewLockBox()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response := struct {
			PrivateKey string `json:"private_key"`
			PublicKey  string `json:"public_key"`
			Address    string `json:"address"`
		}{
			lockbox1.GetPrivateKeyHex(),
			lockbox1.GetPublicKeyHex(),
			lockbox1.GetAddress(),
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseBytes)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}

}

func (vs *Vault) CreateNewVault(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		vault1, err := vault.NewVault()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response := struct {
			PrivateKey string `json:"private_key"`
			PublicKey  string `json:"public_key"`
			Address    string `json:"address"`
		}{
			vault1.GetPrivateKeyHex(),
			vault1.GetPublicKeyHex(),
			vault1.GetAddress(),
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseBytes)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}

}

func (vs *Vault) GetTotalCryptoFromWallet(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Use Set to avoid duplicate headers

	if req.Method == http.MethodGet {
		address := req.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Missing 'address' parameter", http.StatusBadRequest)
			return
		}

		params := url.Values{}
		params.Set("address", address) // Use Set

		blockchainNodeURL, err := url.Parse(vs.BlockchainNodeAddress + "/balance") // Parse the base URL first
		if err != nil {
			http.Error(w, "Invalid blockchain node address", http.StatusInternalServerError)
			return
		}
		blockchainNodeURL.RawQuery = params.Encode() // Add query parameters

		resp, err := http.Get(blockchainNodeURL.String()) // Use the parsed URL
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching balance: %v", err), http.StatusInternalServerError) // More informative error message
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)                                                                // Read the error response body for debugging
			http.Error(w, fmt.Sprintf("Blockchain node returned error: %s", string(body)), resp.StatusCode) // Return the blockchain node's error
			return
		}

		data, err := io.ReadAll(resp.Body) // Use io.ReadAll
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading response body: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK) // Explicitly set status code
		w.Write(data)                // Use w.Write for []byte

	} else {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed) // More appropriate status code
	}
}

func (vs *Vault) SendTxnToTheBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodPost {
		privateKey := req.URL.Query().Get("privateKey")

		dataBs, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Create a new ReadCloser from the byte slice
		bodyReader := io.NopCloser(bytes.NewBuffer(dataBs))

		// Assign the new ReadCloser back to req.Body
		req.Body = bodyReader // Reset req.Body for potential reuse

		// Now unmarshal the data
		var txn1 chain.Transaction
		err = json.Unmarshal(dataBs, &txn1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lockbox, err := lockbox.NewLockBoxFromPrivateKeyHex(privateKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		myTxn := chain.NewTransaction(lockbox.GetAddress(), txn1.To, txn1.Value, []byte(
			""))
		myTxn.Status = constants.PENDING
		newTxn, err := lockbox.GetSignedTxn(*myTxn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newTxnBs, err := json.Marshal(newTxn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// send it to the blockchain
		resp, err := http.Post(vs.BlockchainNodeAddress+"/send_txn", "application/json", bytes.NewBuffer(newTxnBs))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resultBs, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		//Set CORS headers
		SetCorsHeaders(w)

		w.WriteHeader(http.StatusOK)
		w.Write(resultBs)

	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}

}

func (vs *Vault) Start() {
	http.HandleFunc("/vault_balance", vs.GetTotalCryptoFromWallet)
	http.HandleFunc("/create_new_lockbox", vs.CreateNewLockBox)
	http.HandleFunc("/create_new_vault", vs.CreateNewVault)
	http.HandleFunc("/send_signed_txn", vs.SendTxnToTheBlockchain)
	log.Println("Starting wallet server at port:", vs.Port)
	err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(int(vs.Port)), nil)
	if err != nil {
		panic(err)
	}
}

func SetCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Max-Age", "3600")
}
