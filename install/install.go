package install

import (
	"bufio"

	"KNIRVCHAIN/chain"
	"KNIRVCHAIN/utils"
	"log"

	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func prompt(question string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func ConfigureChainPath() (string, error) {
	defaultPath := filepath.Join(utils.UserHomeDir(), "./database/knirv")
	input := prompt(fmt.Sprintf("Enter path to KNIRV Chain data (or press Enter for default): "))

	if input == "" {
		return defaultPath, nil
	}

	return input, nil
}

func DeployChain(path string, ownerAddress string, port string) error {
	dbPath := filepath.Join(path, "knirvdb")
	db, err := chain.NewDBClient(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create new blockchain: %w", err)
	}
	defer db.Close()

	genesisBlock := chain.NewBlock([]byte{}, 0, 0, nil)
	blockchain := chain.NewBlockchain(genesisBlock, "http://localhost:"+port, db)
	err = db.PutIntoDb(blockchain, "http://localhost:"+port)
	if err != nil {
		return fmt.Errorf("failed to put blockchain to db: %w", err)
	}
	log.LogInfo(fmt.Sprintf("Created genesis block with ID: %x", genesisBlock.Hash()))
	return nil
}

func InstallProcess() error {
	fmt.Println("[INFO] running install process")

	path, err := ConfigureChainPath()
	if err != nil {
		return fmt.Errorf("failed to ask for chain path: %w", err)
	}

	ownerAddress := "some_owner_address" // this can be changed
	port := "5000"                       // this can be changed
	err = DeployChain(path, ownerAddress, port)
	if err != nil {
		return fmt.Errorf("failed to create new chain: %w", err)
	}

	return nil
}

func VerifyChainDeployment(path string) error {
	dbPath := filepath.Join(path, "knirv.db")
	db, err := chain.NewDBClient(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create new blockchain: %w", err)
	}
	defer db.Close()
	_, err = db.GetBlockchain("some_address") // You can pass in an address here.
	if err != nil {
		return fmt.Errorf("failed to get blockchain from database: %w", err)
	}

	return nil
}
