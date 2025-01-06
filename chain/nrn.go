package chain

import (
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// NRN struct
type NRN struct {
	name        string
	symbol      string
	totalSupply *big.Int
	maxSupply   *big.Int
	balances    map[common.Address]*big.Int
	owner       common.Address
}

// NewNRN creates a new NRN token
func NewNRN(name string, symbol string, initialSupply *big.Int, maxSupply *big.Int, ownerPrivateKey string) *NRN {
	ownerPrivateKeyECDSA, err := crypto.HexToECDSA(ownerPrivateKey)
	if err != nil {
		log.Fatal("Failed to parse owner private key: ", err)
	}

	ownerAddress := crypto.PubkeyToAddress(ownerPrivateKeyECDSA.PublicKey)

	nrn := &NRN{
		name:        name,
		symbol:      symbol,
		totalSupply: initialSupply,
		maxSupply:   maxSupply,
		balances:    make(map[common.Address]*big.Int),
		owner:       ownerAddress,
	}

	nrn.balances[ownerAddress] = initialSupply
	return nrn
}

// Mint function to mint new tokens
func (n *NRN) Mint(fromPrivateKey string, to common.Address, amount *big.Int) (bool, error) {
	fromAddress := GetAddressFromPrivateKey(fromPrivateKey)
	if fromAddress != n.owner {
		return false, errors.New("only the owner can mint tokens")
	}

	if new(big.Int).Add(n.totalSupply, amount).Cmp(n.maxSupply) > 0 {
		return false, errors.New("minting would exceed max supply")
	}

	if _, exists := n.balances[to]; !exists {
		n.balances[to] = big.NewInt(0)
	}

	n.balances[to] = new(big.Int).Add(n.balances[to], amount)
	n.totalSupply = new(big.Int).Add(n.totalSupply, amount)
	return true, nil
}

// GetBalance gets balance of a given address
func (n *NRN) GetBalance(address common.Address) *big.Int {
	balance, exists := n.balances[address]
	if !exists {
		return big.NewInt(0)
	}
	return balance
}

func (n *NRN) GetTotalSupply() *big.Int {
	return n.totalSupply
}

func (n *NRN) GetOwner() common.Address {
	return n.owner
}

func (n *NRN) Transfer(fromPrivateKey string, to common.Address, amount *big.Int) (bool, error) {
	fromPrivateKeyECDSA, err := crypto.HexToECDSA(fromPrivateKey)
	if err != nil {
		fmt.Println("Failed to parse private key:", err)
		return false, errors.New("failed to parse private key")
	}
	fromAddress := crypto.PubkeyToAddress(fromPrivateKeyECDSA.PublicKey)

	if n.balances[fromAddress].Cmp(amount) < 0 {
		fmt.Println("Insufficient balance")
		return false, errors.New("insufficient balance")
	}
	n.balances[fromAddress] = new(big.Int).Sub(n.balances[fromAddress], amount)
	n.balances[to] = new(big.Int).Add(n.balances[to], amount)
	return true, nil
}

func GeneratePrivateKey() string {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate private key: ", err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	return fmt.Sprintf("%x", privateKeyBytes)

}

func GetAddressFromPrivateKey(privateKey string) common.Address {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	return crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
}
