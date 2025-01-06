package chain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PeerClient struct {
	peerAddress string
}

func (c PeerClient) NewPeerClient(param any) {
	panic("unimplemented")
}

func NewPeerClient(p_address string) *PeerClient {
	return &PeerClient{peerAddress: p_address}
}

func (c *PeerClient) BroadcastTransaction(transaction *Transaction) (bool, error) {
	url := fmt.Sprintf("http://%s/transaction", c.peerAddress)
	data, err := json.Marshal(transaction)
	if err != nil {
		return false, fmt.Errorf("failed to marshal transaction: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return false, fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("POST request failed: %s. Status Code: %d, Body: %s", url, resp.StatusCode, string(body))
	}

	return true, nil
}

func (c *PeerClient) BroadcastBlock(block *Block) (bool, error) {
	url := fmt.Sprintf("http://%s/block", c.peerAddress)
	data, err := json.Marshal(block)
	if err != nil {
		return false, fmt.Errorf("failed to marshal block: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return false, fmt.Errorf("failed to send POST request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("POST request failed: %s. Status Code: %d, Body: %s", url, resp.StatusCode, string(body))
	}
	return true, nil

}

func (c *PeerClient) Close() error {
	return nil
}
