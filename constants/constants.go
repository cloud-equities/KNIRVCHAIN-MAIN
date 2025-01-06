package constants

const (
	BLOCKCHAIN_NAME           = "KNIRVCHAIN"
	HEX_PREFIX                = "0x"
	SUCCESS                   = "success"
	FAILED                    = "failed"
	PENDING                   = "pending"
	MINING_DIFFICULTY         = 5
	MINING_REWARD             = 1200 * DECIMAL
	CURRENCY_NAME             = "nrn"
	DECIMAL                   = 100
	BLOCKCHAIN_ADDRESS        = "KNIRVCHAIN_Faucet"
	BLOCKCHAIN_DB_PATH        = "database/knirv.db"
	BLOCKCHAIN_KEY            = "blockchain_key"
	ADDRESS_PREFIX            = "knirvchain"
	TXN_VERIFICATION_SUCCESS  = "verification_success"
	TXN_VERIFICATION_FAILURE  = "verification_failure"
	BLOCKCHAIN_STATUS         = "RUNNING"
	PEER_BROADCAST_PAUSE_TIME = 1  // In seconds
	PEER_PING_PAUSE_TIME      = 60 // In seconds
	TXN_BROADCAST_PAUSE_TIME  = 1  // In seconds
	FETCH_LAST_N_BLOCKS       = 50
	CONSENSUS_PAUSE_TIME      = 10 // In seconds
)
