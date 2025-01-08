package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"

	"KNIRVCHAIN-MAIN/blockchain"
	"KNIRVCHAIN-MAIN/blockchainserver"
	"KNIRVCHAIN-MAIN/constants"
	"KNIRVCHAIN-MAIN/walletserver"
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
	cfg.PeerBroadcastPauseTime, err1 = strconv.Atoi(os.Getenv("PEER_BROADCAST_PAUSE_TIME"))
	cfg.PeerPingPauseTime, err1 = strconv.Atoi(os.Getenv("PEER_PING_PAUSE_TIME"))
	cfg.TxnBroadcastPauseTime, err1 = strconv.Atoi(os.Getenv("TXN_BROADCAST_PAUSE_TIME"))
	cfg.FetchLastNBlocks, err1 = strconv.Atoi(os.Getenv("FETCH_LAST_N_BLOCKS"))
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

	// Load configuration from .env
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	chainCmdSet := flag.NewFlagSet("chain", flag.ExitOnError)
	walletCmdSet := flag.NewFlagSet("wallet", flag.ExitOnError)

	chainPort := chainCmdSet.Uint64("port", cfg.Port, "HTTP port to launch our blockchain server")
	chainMiner := chainCmdSet.String(cfg.MinersAddress, "", "Miners address to credit mining reward")
	remoteNode := chainCmdSet.String("remote_node", "", "Remote Node from where the blockchain will be synced")

	walletPort := walletCmdSet.Uint64("port", 8080, "HTTP port to launch our wallet server")
	blockchainNodeAddress := walletCmdSet.String("node_address", "http://127.0.0.1:5000", "Blockchain node address for the wallet gateway")

	if len(os.Args) < 2 {
		fmt.Println("Error:Expected chain or wallet subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "chain":
		var wg sync.WaitGroup

		chainCmdSet.Parse(os.Args[2:])
		if chainCmdSet.Parsed() {
			if *chainMiner == "" || chainCmdSet.NFlag() == 0 {
				fmt.Println("Usage of chain subcommand: ")
				chainCmdSet.PrintDefaults()
				os.Exit(1)
			}

			if *remoteNode == "" {

				genesisBlock := blockchain.NewBlock("0x0", 0, 0)
				blockchain1 := blockchain.NewBlockchain(*genesisBlock, "http://127.0.0.1:"+strconv.Itoa(int(*chainPort)))
				blockchain1.Peers[blockchain1.Address] = true
				bcs := blockchainserver.NewBlockchainServer(*chainPort, blockchain1)
				wg.Add(4)

				for _, peerAddress := range cfg.PeerAddresses {
					pm.AddPeer(peerAddress) // Assuming AddPeer handles connecting to new peers
				}
				go bcs.Start()
				go bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner)
				go bcs.BlockchainPtr.DialAndUpdatePeers()
				go bcs.BlockchainPtr.RunConsensus()
				wg.Wait()
			} else {
				blockchain1, err := blockchain.SyncBlockchain(*remoteNode)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				blockchain2 := blockchain.NewBlockchainFromSync(blockchain1, "http://127.0.0.1:"+strconv.Itoa(int(*chainPort)))
				blockchain2.Peers[blockchain2.Address] = true
				bcs := blockchainserver.NewBlockchainServer(*chainPort, blockchain2)
				wg.Add(4)
				go bcs.Start()
				go bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner)
				go bcs.BlockchainPtr.DialAndUpdatePeers()
				go bcs.BlockchainPtr.RunConsensus()
				wg.Wait()
			}

		}
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
