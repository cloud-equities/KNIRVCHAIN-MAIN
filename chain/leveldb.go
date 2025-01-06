package chain

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LevelDB struct {
	Client *leveldb.DB
}

func NewDBClient(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}
	return &LevelDB{Client: db}, nil
}

func (db *LevelDB) SaveLastBlock(block interface{}, address string) error {
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return err
	}
	err = db.Client.Put([]byte(address+"last_block"), blockBytes, &opt.WriteOptions{})
	if err != nil {
		return fmt.Errorf("failed to put data into database with key last_block: %w", err)
	}
	return nil
}

func (db *LevelDB) LoadLastBlock(address string) (interface{}, error) {
	data, err := db.Client.Get([]byte(address+"last_block"), &opt.ReadOptions{})
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("no block found: %w", err)
		}
		return nil, fmt.Errorf("failed to get data from database with key last_block: %w", err)
	}

	block := new(Block)
	err = json.Unmarshal(data, block)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return block, nil
}

func (db *LevelDB) KeyExists(address string) (bool, error) {
	_, err := db.Client.Get([]byte(address), &opt.ReadOptions{})
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get data from database with key blockchain: %w", err)

	}
	return true, nil
}

func (db *LevelDB) GetBlockchain(address string) (interface{}, error) {
	data, err := db.Client.Get([]byte(address), &opt.ReadOptions{})
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("no blockchain found: %w", err)
		}
		return nil, fmt.Errorf("failed to get data from database with key blockchain: %w", err)
	}

	blockchain := new(BlockchainStruct)
	err = json.Unmarshal(data, blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return blockchain, nil
}

func (db *LevelDB) PutIntoDb(blockchain interface{}, address string) error {
	data, err := json.Marshal(blockchain)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = db.Client.Put([]byte(address), data, &opt.WriteOptions{})
	if err != nil {
		return fmt.Errorf("failed to put data into database with key blockchain: %w", err)
	}
	return nil
}

func (db *LevelDB) Close() error {
	return db.Client.Close()
}
