package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	//"sync"
	"time"

	"github.com/joho/godotenv"

	"KNIRVCHAIN-MAIN/block"
	"KNIRVCHAIN-MAIN/blockchain"
	"KNIRVCHAIN-MAIN/blockchainserver"
	"KNIRVCHAIN-MAIN/consensus"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/events"
	"KNIRVCHAIN-MAIN/peerManager"
	"KNIRVCHAIN-MAIN/transaction"
	"KNIRVCHAIN-MAIN/walletserver"
)

const (
	minersAddressFlag = "miners_address" // Define a constant for the flag name
)

type Config struct {
	Port                   uint64
	MinersAddress          string
	DatabasePath           string
	BlockchainName         string
	HexPrefix              string
	Success                bool
	Failed                 bool
	Pending                string
	MiningDifficulty       int
	MiningReward           int64
	CurrencyName           string
	Decimal                int
	BlockchainAddress      string
	BlockchainDbPath       string
	BlockchainKey          string
	AddressPrefix          string
	TxnVerificationSuccess string
	TxnVerificationFailure string
	BlockchainStatus       string
	PeerBroadcastPauseTime int
	PeerPingPauseTime      int
	TxnBroadcastPauseTime  int
	FetchLastNBlocks       int
	ConsensusPauseTime     int
	PeerAddresses          []string
}

func loadConfig() (*Config, error) {
	//Load .env file
	err := godotenv.Load()

	if os.Getenv("MINING_DIFFICULTY") == "" {
		return nil, fmt.Errorf("MINING_DIFFICULTY is not set")
	}
	if os.Getenv("MINING_REWARD") == "" {
		return nil, fmt.Errorf("MINING_REWARD is not set")
	}
	if os.Getenv("DECIMAL") == "" {
		return nil, fmt.Errorf("DECIMAL is not set")
	}
	if os.Getenv("CONSENSUS_PAUSE_TIME") == "" {

		return nil, fmt.Errorf("CONSENSUS_PAUSE_TIME is not set")
	}

	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	cfg := &Config{}
	var err1 error
	cfg.Port, err1 = strconv.ParseUint(os.Getenv("PORT"), 10, 64)
	if err1 != nil {
		return nil, fmt.Errorf("error parsing PORT: %w", err1)
	}
	cfg.MinersAddress = os.Getenv("MINERS_ADDRESS")
	cfg.DatabasePath = os.Getenv("DATABASE_PATH")
	cfg.BlockchainName = os.Getenv("BLOCKCHAIN_NAME")
	cfg.HexPrefix = os.Getenv("HEX_PREFIX")
	cfg.Success = os.Getenv("SUCCESS") == "true"
	cfg.Failed = os.Getenv("FAILED") == "true"
	cfg.Pending = os.Getenv("PENDING")
	cfg.MiningDifficulty, err1 = strconv.Atoi(os.Getenv("MINING_DIFFICULTY"))
	if err1 != nil {
		return nil, fmt.Errorf("error parsing MINING_DIFFICULTY: %w", err1)
	}
	cfg.MiningReward, err1 = strconv.ParseInt(os.Getenv("MINING_REWARD"), 10, 64)
	if err1 != nil {
		return nil, fmt.Errorf("error parsing MINING_REWARD: %w", err1)
	}
	cfg.CurrencyName = os.Getenv("CURRENCY_NAME")
	cfg.Decimal, err1 = strconv.Atoi(os.Getenv("DECIMAL"))
	if err1 != nil {
		return nil, fmt.Errorf("error parsing DECIMAL: %w", err1)
	}
	cfg.BlockchainAddress = os.Getenv("BLOCKCHAIN_ADDRESS")
	cfg.BlockchainDbPath = os.Getenv("BLOCKCHAIN_DB_PATH")
	cfg.BlockchainKey = os.Getenv("BLOCKCHAIN_KEY")
	cfg.AddressPrefix = os.Getenv("ADDRESS_PREFIX")
	cfg.TxnVerificationSuccess = os.Getenv("TXN_VERIFICATION_SUCCESS")
	cfg.TxnVerificationFailure = os.Getenv("TXN_VERIFICATION_FAILURE")
	cfg.BlockchainStatus = os.Getenv("BLOCKCHAIN_STATUS")
	//cfg.PeerBroadcastPauseTime, err1 = strconv.Atoi(os.Getenv("PEER_BROADCAST_PAUSE_TIME"))
	//cfg.PeerPingPauseTime, err1 = strconv.Atoi(os.Getenv("PEER_PING_PAUSE_TIME"))
	//cfg.TxnBroadcastPauseTime, err1 = strconv.Atoi(os.Getenv("TXN_BROADCAST_PAUSE_TIME"))
	//cfg.FetchLastNBlocks, err1 = strconv.Atoi(os.Getenv("FETCH_LAST_N_BLOCKS"))
	cfg.ConsensusPauseTime, err1 = strconv.Atoi(os.Getenv("CONSENSUS_PAUSE_TIME"))
	if err1 != nil {
		return nil, fmt.Errorf("error parsing an integer config value: %w", err1)
	}
	//cfg.PeerAddresses = strings.Split(os.Getenv("PEER_ADDRESSES"), ",")
	peerString := os.Getenv("PEER_ADDRESSES")
	if peerString != "" {
		cfg.PeerAddresses = strings.Split(peerString, ",")
	}

	// Validate loaded values (e.g., check for empty strings)
	if cfg.BlockchainDbPath == "" || cfg.MinersAddress == "" {
		return nil, fmt.Errorf("missing required configuration values")
	}

	return cfg, nil
}

func init() {
	log.SetPrefix(constants.BLOCKCHAIN_NAME + ":")
}

func main() {

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Println("Error loading configuration:", err)
		os.Exit(1)
	}

	// Define command-line flags
	chainCmdSet := flag.NewFlagSet("chain", flag.ExitOnError)
	walletCmdSet := flag.NewFlagSet("wallet", flag.ExitOnError)

	chainPort := chainCmdSet.Uint64("port", cfg.Port, "HTTP port for blockchain server")
	chainMiner := chainCmdSet.String(minersAddressFlag, "", "Miner's address")
	remoteNode := chainCmdSet.String("remote_node", "", "Remote node for syncing")

	walletPort := walletCmdSet.Uint64("port", 8080, "HTTP port for wallet server")
	blockchainNodeAddress := walletCmdSet.String("node_address", "http://127.0.0.1:5001", "Blockchain node address")

	// Check for subcommand
	if len(os.Args) < 2 {
		fmt.Println("Error: Expected 'chain' or 'wallet' subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "chain":
		//var wg sync.WaitGroup
		chainCmdSet.Parse(os.Args[2:])

		// Validate flags
		if chainCmdSet.Parsed() {
			if *chainMiner == "" || chainCmdSet.NFlag() == 0 {
				fmt.Println("Usage of chain subcommand:")
				chainCmdSet.PrintDefaults()
				os.Exit(1)
			}
		}

		startMining := make(chan bool)
		startConsensus := make(chan bool)
		stopMining := make(chan bool)
		miningStopped := make(chan bool)

		var bcs *blockchainserver.BlockchainServer
		var consensusMgr *consensus.ConsensusManager
		var pm *peerManager.PeerManager
		var blockchain1 *blockchain.BlockchainStruct

		blockAddedChan := make(chan events.BlockAddedEvent)
		transactionAddedChan := make(chan events.TransactionAddedEvent)
		pm = consensus.GetPeerManager(blockAddedChan, transactionAddedChan)

		if *remoteNode == "" {
			genesisBlock := block.NewBlock("0x0", 0, 0)

			pm.Address = "http://127.0.0.1:" + strconv.Itoa(int(*chainPort))

			pm.Broadcaster = peerManager.PeerTransactionBroadcaster{PeerManager: pm}
			blockchain1 = blockchain.NewBlockchain(*genesisBlock, pm.Address, &pm.Broadcaster, pm)
			blockchain1.Peers[blockchain1.Address] = true
			bcs = blockchainserver.NewBlockchainServer(*chainPort, blockchain1)
			consensusMgr = consensus.NewConsensusManager(blockchain1, pm)
			go func() { // This goroutine MUST start AFTER blockchain1 is initialized
				for event := range blockchain1.TransactionAdded {
					pm.UpdateTransactionPool([]*transaction.Transaction{event.Transaction})
				}
			}()

		} else {
			//genesisBlock := block.NewBlock("0x0", 0, 0)
			pm.Address = "http://127.0.0.1:" + strconv.Itoa(int(*chainPort))
			pm.Broadcaster = peerManager.PeerTransactionBroadcaster{PeerManager: pm}

			//blockchain1 = blockchain.NewBlockchain(*genesisBlock, pm.Address, &pm.Broadcaster, pm)

			// if *remoteNode != ""
			remotePeerManager, err := peerManager.SyncBlockchain(*remoteNode)
			if err != nil {
				log.Println(err)
			}

			blockchain1 = blockchain.NewBlockchainFromSync(remotePeerManager.Blocks, pm.Address, &pm.Broadcaster, pm)
			blockchain1.Peers[blockchain1.Address] = true
			bcs = blockchainserver.NewBlockchainServer(*chainPort, blockchain1)
			consensusMgr = consensus.NewConsensusManager(blockchain1, pm)

			go func() { // This goroutine MUST start AFTER blockchain1 is initialized
				for event := range blockchain1.TransactionAdded {
					pm.UpdateTransactionPool([]*transaction.Transaction{event.Transaction})
				}
			}()

		}

		//wg.Add(1) // Wait for the server to start
		bcs.Start()

		//wg.Wait() // Server is running now

		go pm.StartListening() // Start peer management AFTER server is up
		go pm.DialAndUpdatePeers()

		go func() {
			<-startMining
			bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner, stopMining, miningStopped)
		}()

		go func() {
			<-startConsensus
			consensusMgr.RunConsensus(startMining)
		}()

		// Example: Trigger mining and consensus based on some condition or delay.
		time.Sleep(5 * time.Second)
		startMining <- true
		startConsensus <- true

	case "wallet":
		walletCmdSet.Parse(os.Args[2:])
		if walletCmdSet.Parsed() {
			if walletCmdSet.NFlag() == 0 {
				fmt.Println("Usage of wallet subcommand: ")
				walletCmdSet.PrintDefaults()
				os.Exit(1)
			}

			ws := walletserver.NewWalletServer(*walletPort, *blockchainNodeAddress)
			ws.Start()
		}
	default:
		fmt.Println("Error:Expected chain or wallet subcommand")
		os.Exit(1)
	}
}
