package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// NFT represents a Non-Fungible Token
type NFT struct {
	ID       string                 `json:"id"`
	Data     []byte                 `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
	Owner    string                 `json:"owner"`
	MintedAt time.Time              `json:"minted_at"`
}

// NewNFT creates a new NFT
func NewNFT(data []byte, metadata map[string]interface{}, owner string) (*NFT, error) {
	timestamp := time.Now()
	nft := &NFT{
		ID:       generateID(data),
		Data:     data,
		Metadata: metadata,
		Owner:    owner,
		MintedAt: timestamp,
	}
	return nft, nil
}

// generateID generates a unique ID for the NFT
func generateID(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

func (b *NFT) GetNRNByNFTID(identifier string) (any, any) {
	panic("unimplemented")
}

// Serialize NFT to byte array
func (nft *NFT) Serialize() ([]byte, error) {
	data, err := json.Marshal(nft)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize nft: %w", err)
	}
	return data, nil
}

// DeserializeNFT recreates NFT from byte array
func DeserializeNFT(data []byte) (*NFT, error) {
	var nft NFT
	err := json.Unmarshal(data, &nft)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize nft: %w", err)
	}
	return &nft, nil
}
