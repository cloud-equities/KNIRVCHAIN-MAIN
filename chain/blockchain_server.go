package chain

import (
	"constants"
	"encoding/json"
	"fmt"
	"html"
	"io"
	knirvlog "log"
	"net/http"
	"strconv"
)

// Struct for your Server.
type BlockchainServer struct {
	Port          uint64            `json:"port"`
	BlockchainPtr *BlockchainStruct `json:"blockchain"`
	ChainAddress  string            `json:"chain_address"` // Keep the address, for now but not as implementation parameter in methods (when possible) to reduce implicit requirements from your methods.
}

func NewBlockchainServer(port uint64, blockchainPtr *BlockchainStruct, chainAddress string) *BlockchainServer {
	bcs := new(BlockchainServer)
	bcs.Port = port
	bcs.BlockchainPtr = blockchainPtr // Pass existing blockchain to avoid new blockchain methods that use that same parameter scope
	bcs.ChainAddress = chainAddress   // set it explicitly for a proper reference

	return bcs
}

// Implements data formatting of HTML components, with a known method to ensure valid JSON and HTML generation with error handling if anything occurs that can generate invalid values
func (bcs *BlockchainServer) GetBlockchain(w http.ResponseWriter, req *http.Request) { // Use struct for parameter implementation for methods where they require struct based information such as data structs with JSON attributes for a known structure when retrieving the information via our interfaces using helper methods such as `PutIntoDb()` for persisting test output from API servers.
	w.Header().Add("Content-Type", "application/json") // Set Header with expected http method to provide feedback of valid or incorrect parameters during API access verification for test cases
	if req.Method == http.MethodGet {                  // ensure correct type has been provided.
		io.WriteString(w, bcs.BlockchainPtr.ToJson()) // now able to pass correct types, when returning data for your program using json marshalling that has already been proven to work in tests using proper type declaration with testing suite.
	} else {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed) // Provide detailed status output using this newly implemented methods.
	}
}

func (bcs *BlockchainServer) GetBalance(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method != http.MethodGet {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}
	addr := req.URL.Query().Get("KNIRVChainAddress") // verify request object types are properly implemented before accessing, the data and their value.
	x := struct {
		Balance uint64 `json:"balance"`
	}{
		bcs.BlockchainPtr.CalculateTotalCrypto(addr),
	}

	mBalance, err := json.Marshal(x) // convert data to valid type for API transport
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	io.WriteString(w, string(mBalance))
}

func (bcs *BlockchainServer) GetAllNonRewardedTxns(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // Ensure errors types conform to methods.
		return
	}
	txnList := bcs.BlockchainPtr.GetAllTxns() // use proper interface to correctly access blockchain methods for your implementation structure.
	byteSlice, err := json.Marshal(txnList)   // test types using method signatures that produce these struct data, so type requirements do not cause any conflicts when passing this data.
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return // proper error handling.
	}
	io.WriteString(w, string(byteSlice))
}

func (bcs *BlockchainServer) SendTxnToTheBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json") // headers always set on top for all implementations of new handlers or services for type specifications in methods.
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // Return standard error, which all tests are looking for from the previous tests.
		return
	}

	request, err := io.ReadAll(req.Body) // use type verification. and return if any issues have been reported in all resource access steps during execution with specific message related to scope of implementation of data structure that did not comply with methods using type signatures.
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer req.Body.Close() // close on method return after function call, instead of keeping methods alive until an interrupt forces it shut, which you do not intend for.

	var newTxn Transaction // test structs using consistent parameters

	err = json.Unmarshal(request, &newTxn) // Unmarshal request to expected structure data
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // log unexpected result and return to a bad result when a request type does not comply with the intended signature for struct values for objects using testing methods, where all methods from the core code should handle errors and data processing of required types with those methods using structs when accessing project logic data and system configuration resources for test suites.
		return                                            // always returns value when bad.
	}
	go bcs.BlockchainPtr.AddTransactionToTransactionPool(&newTxn)

	io.WriteString(w, newTxn.ToJson()) // write to new object using interface pattern for consistent design
}

func CheckStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}
	io.WriteString(w, constants.BLOCKCHAIN_STATUS)
}

func (bcs *BlockchainServer) SendPeersList(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json") // correct header.
	if req.Method != http.MethodPost {                 // enforce all required data when performing these test operations or implementation features that may exist in more than one project scope, and or type during development, for reuse of implementation specific types that adhere to testing requirements, and project specific constraints, during method testing verification to ensure components that use different parameters all conform and implement the project design goals using correct methods during design implementations.
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed) // make sure they only are called in a `Post` request or will send that back with an intended output of errors, using built in test validation systems provided by go compiler to correctly enforce parameters on your struct components using these libraries.
		return
	}
	peersMap, err := io.ReadAll(req.Body) // Use consistent naming scheme for error return value, for better control over project output using known variable naming that indicates if any methods has returned a faulty execution parameter from a method return.
	if err != nil {
		knirvlog.LogError("Error reading Peers:", err) // implement logs consistently so when these methods need debugging you will have appropriate outputs to trace variables back to and perform error correction by isolating problem variables during debug.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return // proper use of try and catch to make all tests and methods resilient to errors with system operation and object usage of data or their interfaces.
	}
	defer req.Body.Close() // you are passing body here and may need to close in future testing so must do correct http method logic to create tests.
	var peersList map[string]bool
	err = json.Unmarshal(peersMap, &peersList)
	if err != nil { // Always verify inputs data before attempting to do complex logic calls using them
		knirvlog.LogError("Error unmarshalling Peers:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return // proper output of the errors for analysis during troubleshooting when using libraries like this for http.Response that it requires consistent checks before running its internal code methods that depend on that http response data from server or another test method.
	}
	go bcs.BlockchainPtr.DialAndUpdatePeers()
	res := map[string]string{}  // using structs when accessing code methods are good since the scope and expected methods are available by defining the parameter explicitly as this particular type (which must follow data implementation structures and not from assumptions about data. that is what interfaces are supposed to handle)
	res["status"] = "success"   // ensure a defined return that does not cause an error from any type conversion issues for your implementations, or any data struct related parameters for this method implementation to verify expected test case functionality at the boundary conditions of method scopes in projects during execution.
	x, err := json.Marshal(res) // if testing passes this object and its related calls, should continue to perform correct execution.
	if err != nil {
		knirvlog.LogError("Error while marshalling the http Response to be sent:", err)
		http.Error(w, err.Error(), http.StatusBadRequest) // add specific log output to tell where, and why a method has an exception in their functionality for system level method logging (not user output which would implement code like using `fmt.Print` to display for humans).
		return
	}
	io.WriteString(w, string(x)) // Write using type interface when making method calls
}
func (bcs *BlockchainServer) FetchLastNBlocks(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json") // verify headers that data follows format or methods signatures from an object struct by explicitly making sure the implementation does not crash or lead to logic errors by unexpected inputs or parameters in tests.
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	blocks := bcs.BlockchainPtr.Blocks
	blockchain1 := new(BlockchainStruct) // type definitions allow for more testable data.
	if len(blocks) < constants.FETCH_LAST_N_BLOCKS {
		blockchain1.Blocks = blocks // check if less data in our known boundary constraints are passed.
	} else {
		blockchain1.Blocks = blocks[len(blocks)-constants.FETCH_LAST_N_BLOCKS:] // Ensure logic operations such as indexing a resource, only do it if the implementation or code type requirement are passed or if not create appropriate testing messages so that errors are resolved to correct parameter types for use in tests.
	}
	byte_output, err := json.Marshal(blockchain1) // proper usage of library, as declared to be of interface value of json.Marshall, since it also must satisfy testing types
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return // Implement all safety and error check, with detailed exception handlers. and logging where possible to keep debug to a minimum in next test implementation passes or build phases.

	}
	io.WriteString(w, string(byte_output))

}
func (bcs *BlockchainServer) Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // Test method response data, for data display verification and functionality testing by looking up parameters through these structs with local tests.
		params := html.DashboardParams{
			ChainAddress:  bcs.ChainAddress,
			WalletAddress: "testingID",            // Implement correct wallet method access for a known testing state before you start calling real wallets for core authentication functions.
			Owner:         "Owner for Test Mode.", // make implementation code for this, to set for every new server that gets called for correct variable use with data struct properties.
			Message:       "Connected to the KNIRVCHAIN.",
		}

		err := html.Dashboard(w, params)
		if err != nil { // all parameters with structs need error messages
			http.Error(w, err.Error(), http.StatusBadRequest)                                 // Always use correct http responses that reflect methods being used with struct properties or methods for correct testing.
			knirvlog.LogError("Failed to load initial server response on default '/': ", err) // check logs to verify tests in context.
		}
	})

	http.HandleFunc("/balance", bcs.GetBalance) // testing the use of proper url parameters with path components
	http.HandleFunc("/get_all_non_rewarded_txns", bcs.GetAllNonRewardedTxns)
	http.HandleFunc("/send_txn", bcs.SendTxnToTheBlockchain)
	http.HandleFunc("/send_peers_list", bcs.SendPeersList)
	http.HandleFunc("/check_status", CheckStatus)
	http.HandleFunc("/fetch_last_n_blocks", bcs.FetchLastNBlocks)

	knirvlog.LogInfo(fmt.Sprintf("Launching webserver at port : %d", bcs.Port)) // Implement Logging properly here in this specific code format so you do not repeat error conditions of `Println` system methods being mixed with code calls for implementation of your own local logging framework implementation in go, that has proper test methods
	err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(int(bcs.Port)), nil)
	if err != nil {
		panic(err) // stop testing flow for major problems for data structures with specific implementations for core logic testing purposes.
	}
}
