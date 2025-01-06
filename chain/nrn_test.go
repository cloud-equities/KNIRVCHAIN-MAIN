package chain

import (
	"math/big"
	"testing"
)

func TestNRN(t *testing.T) {
	ownerPrivateKey := GeneratePrivateKey()
	ownerAddress := GetAddressFromPrivateKey(ownerPrivateKey)
	initialSupply := big.NewInt(1000)
	maxSupply := big.NewInt(10000)

	nrn := NewNRN("TestToken", "TT", initialSupply, maxSupply, ownerPrivateKey)

	// Test initial state
	if nrn.GetTotalSupply().Cmp(initialSupply) != 0 {
		t.Errorf("Initial total supply is incorrect. Expected %v, got %v", initialSupply, nrn.GetTotalSupply())
	}

	if nrn.GetBalance(ownerAddress).Cmp(initialSupply) != 0 {
		t.Errorf("Initial owner balance is incorrect. Expected %v, got %v", initialSupply, nrn.GetBalance(ownerAddress))
	}

	if nrn.GetOwner() != ownerAddress {
		t.Errorf("Owner address is incorrect")
	}

	// Test minting
	recipientPrivateKey := GeneratePrivateKey()
	recipientAddress := GetAddressFromPrivateKey(recipientPrivateKey)

	mintAmount := big.NewInt(500)

	_, err := nrn.Mint(ownerPrivateKey, recipientAddress, mintAmount)
	if err != nil {
		t.Fatalf("Minting failed: %v", err)
	}

	expectedTotalSupply := new(big.Int).Add(initialSupply, mintAmount)
	if nrn.GetTotalSupply().Cmp(expectedTotalSupply) != 0 {
		t.Errorf("Total supply is incorrect after minting. Expected %v, got %v", expectedTotalSupply, nrn.GetTotalSupply())
	}
	if nrn.GetBalance(recipientAddress).Cmp(mintAmount) != 0 {
		t.Errorf("Recipient balance is incorrect after minting. Expected %v, got %v", mintAmount, nrn.GetBalance(recipientAddress))
	}
	if nrn.GetBalance(ownerAddress).Cmp(initialSupply) != 0 {
		t.Errorf("Owner balance is incorrect after minting. Expected %v, got %v", initialSupply, nrn.GetBalance(ownerAddress))
	}

	// Test minting fails when not owner
	_, err = nrn.Mint(recipientPrivateKey, recipientAddress, mintAmount)
	if err == nil {
		t.Fatalf("Minting should fail when not owner")
	}

	// Test minting over max supply
	remaining := new(big.Int).Sub(maxSupply, nrn.GetTotalSupply())

	mintAmountOverMax := new(big.Int).Add(remaining, big.NewInt(1))
	_, err = nrn.Mint(ownerPrivateKey, recipientAddress, mintAmountOverMax)
	if err == nil {
		t.Fatalf("Minting over max supply should fail")
	}

	// Test transfer
	transferAmount := big.NewInt(200)

	_, err = nrn.Transfer(ownerPrivateKey, recipientAddress, transferAmount)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	expectedOwnerBalanceAfterTransfer := new(big.Int).Sub(initialSupply, transferAmount)
	if nrn.GetBalance(ownerAddress).Cmp(expectedOwnerBalanceAfterTransfer) != 0 {
		t.Errorf("Owner balance is incorrect after transfer. Expected %v, got %v", expectedOwnerBalanceAfterTransfer, nrn.GetBalance(ownerAddress))
	}

	expectedRecipientBalanceAfterTransfer := new(big.Int).Add(mintAmount, transferAmount)
	if nrn.GetBalance(recipientAddress).Cmp(expectedRecipientBalanceAfterTransfer) != 0 {
		t.Errorf("Recipient balance is incorrect after transfer. Expected %v, got %v", expectedRecipientBalanceAfterTransfer, nrn.GetBalance(recipientAddress))
	}

	// Test transfer fails when insufficient balance
	transferAmountTooMuch := big.NewInt(1000000)

	_, err = nrn.Transfer(ownerPrivateKey, recipientAddress, transferAmountTooMuch)
	if err == nil {
		t.Fatalf("Transfer should fail when insufficient funds")
	}
}
