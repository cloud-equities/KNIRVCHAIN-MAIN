package vault

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"

	constants "github.com/cloud-equities/KNIRVCHAIN/constants"

	"github.com/cloud-equities/KNIRVCHAIN-chain"
)

type LockBox struct {
	PrivateKey *ecdsa.PrivateKey `json:"private_key"`
	PublicKey  *ecdsa.PublicKey  `json:"public_key"`
}

func NewLockBox() (*LockBox, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	lockBox := new(LockBox)
	lockBox.PrivateKey = privateKey
	lockBox.PublicKey = &privateKey.PublicKey

	return lockBox, nil
}

func NewLockBoxFromPrivateKeyHex(privateKeyHex string) *LockBox {
	pk := privateKeyHex[2:]
	d := new(big.Int)
	d.SetString(pk, 16)

	var npk ecdsa.PrivateKey
	npk.D = d
	npk.PublicKey.Curve = elliptic.P256()
	npk.PublicKey.X, npk.PublicKey.Y = npk.PublicKey.Curve.ScalarBaseMult(d.Bytes())

	lockBox := new(LockBox)
	lockBox.PrivateKey = &npk
	lockBox.PublicKey = &npk.PublicKey

	return lockBox
}

func (w *LockBox) GetPrivateKeyHex() string {
	return fmt.Sprintf("0x%x", w.PrivateKey.D)
}

func (w *LockBox) GetPublicKeyHex() string {
	return fmt.Sprintf("0x%x%x", w.PublicKey.X, w.PublicKey.Y)
}

func (w *LockBox) GetAddress() string {
	hash := sha256.Sum256([]byte(w.GetPublicKeyHex()[2:]))
	hex := fmt.Sprintf("%x", hash[:])
	block_chainaddress := constants.ADDRESS_PREFIX + hex[len(hex)-40:]
	return block_chainaddress
}

func (w *LockBox) GetSignedTxn(unsignedTxn chain.Transaction) (*chain.Transaction, error) {
	bs, err := json.Marshal(unsignedTxn)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(bs)

	sig, err := ecdsa.SignASN1(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return nil, err
	}

	var signedTxn chain.Transaction
	signedTxn.From = unsignedTxn.From
	signedTxn.To = unsignedTxn.To
	signedTxn.Data = unsignedTxn.Data
	signedTxn.Status = unsignedTxn.Status
	signedTxn.Value = unsignedTxn.Value
	signedTxn.Timestamp = unsignedTxn.Timestamp
	signedTxn.TransactionHash = unsignedTxn.TransactionHash
	// new fields
	signedTxn.Signature = sig
	signedTxn.PublicKey = w.GetPublicKeyHex()

	return &signedTxn, nil
}
