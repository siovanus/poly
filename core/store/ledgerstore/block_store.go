/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"io"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/serialization"
	scom "github.com/polynetwork/poly/core/store/common"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/types"
)

// Block store save the data of block & transaction
type BlockStore struct {
	enableCache bool                       //Is enable lru cache
	dbDir       string                     //The path of store file
	cache       *BlockCache                //The cache of block, if have.
	store       *leveldbstore.LevelDBStore //block store handler
}

// NewBlockStore return the block store instance
func NewBlockStore(dbDir string, enableCache bool) (*BlockStore, error) {
	var cache *BlockCache
	var err error
	if enableCache {
		cache, err = NewBlockCache()
		if err != nil {
			return nil, fmt.Errorf("NewBlockCache error %s", err)
		}
	}

	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	blockStore := &BlockStore{
		dbDir:       dbDir,
		enableCache: enableCache,
		store:       store,
		cache:       cache,
	}
	return blockStore, nil
}

// NewBatch start a commit batch
func (this *BlockStore) NewBatch() {
	this.store.NewBatch()
}

// SaveBlock persist block to store
func (this *BlockStore) SaveBlock(block *types.Block) error {
	if this.enableCache {
		this.cache.AddBlock(block)
	}

	blockHeight := block.Header.Height
	err := this.SaveHeader(block)
	if err != nil {
		return fmt.Errorf("SaveHeader error %s", err)
	}
	for _, tx := range block.Transactions {
		err = this.SaveTransaction(tx, blockHeight)
		if err != nil {
			txHash := tx.Hash()
			return fmt.Errorf("SaveTransaction block height %d tx %s err %s", blockHeight, txHash.ToHexString(), err)
		}
	}
	return nil
}

// ContainBlock return the block specified by block hash save in store
func (this *BlockStore) ContainBlock(blockHash common.Uint256) (bool, error) {
	if this.enableCache {
		if this.cache.ContainBlock(blockHash) {
			return true, nil
		}
	}
	key := this.getHeaderKey(blockHash)
	_, err := this.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBlock return block by block hash
func (this *BlockStore) GetBlock(blockHash common.Uint256) (*types.Block, error) {
	var block *types.Block
	if this.enableCache {
		block = this.cache.GetBlock(blockHash)
		if block != nil {
			return block, nil
		}
	}
	header, txHashes, err := this.loadHeaderWithTx(blockHash)
	if err != nil {
		return nil, err
	}
	txList := make([]*types.Transaction, 0, len(txHashes))
	for _, txHash := range txHashes {
		tx, _, err := this.GetTransaction(txHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction %s error %s", txHash.ToHexString(), err)
		}
		if tx == nil {
			return nil, fmt.Errorf("cannot get transaction %s", txHash.ToHexString())
		}
		txList = append(txList, tx)
	}
	block = &types.Block{
		Header:       header,
		Transactions: txList,
	}
	return block, nil
}

func (this *BlockStore) loadHeaderWithTx(blockHash common.Uint256) (*types.Header, []common.Uint256, error) {
	key := this.getHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, nil, err
	}
	source := common.NewZeroCopySource(value)
	header := new(types.Header)
	err = header.Deserialization(source)
	if err != nil {
		return nil, nil, err
	}
	txSize, eof := source.NextUint32()
	if eof {
		return nil, nil, io.ErrUnexpectedEOF
	}
	txHashes := make([]common.Uint256, 0, int(txSize))
	for i := uint32(0); i < txSize; i++ {
		txHash, eof := source.NextHash()
		if eof {
			return nil, nil, io.ErrUnexpectedEOF
		}
		txHashes = append(txHashes, txHash)
	}
	return header, txHashes, nil
}

// SaveHeader persist block header to store
func (this *BlockStore) SaveHeader(block *types.Block) error {
	blockHash := block.Hash()
	key := this.getHeaderKey(blockHash)
	sink := common.NewZeroCopySink(nil)
	block.Header.Serialization(sink)
	sink.WriteUint32(uint32(len(block.Transactions)))
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		sink.WriteHash(txHash)
	}
	this.store.BatchPut(key, sink.Bytes())
	return nil
}

// GetHeader return the header specified by block hash
func (this *BlockStore) GetHeader(blockHash common.Uint256) (*types.Header, error) {
	if this.enableCache {
		block := this.cache.GetBlock(blockHash)
		if block != nil {
			return block.Header, nil
		}
	}
	return this.loadHeader(blockHash)
}

func (this *BlockStore) loadHeader(blockHash common.Uint256) (*types.Header, error) {
	key := this.getHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(value)
	header := new(types.Header)
	err = header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// GetCurrentBlock return the current block hash and current block height
func (this *BlockStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := this.getCurrentBlockKey()
	data, err := this.store.Get(key)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Uint256{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

// SaveCurrentBlock persist the current block height and current block hash to store
func (this *BlockStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) error {
	key := this.getCurrentBlockKey()
	value := bytes.NewBuffer(nil)
	blockHash.Serialize(value)
	serialization.WriteUint32(value, height)
	this.store.BatchPut(key, value.Bytes())
	return nil
}

// GetBlockHash return block hash by block height
func (this *BlockStore) GetBlockHash(height uint32) (common.Uint256, error) {
	key := this.getBlockHashKey(height)
	value, err := this.store.Get(key)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	blockHash, err := common.Uint256ParseFromBytes(value)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return blockHash, nil
}

// SaveBlockHash persist block height and block hash to store
func (this *BlockStore) SaveBlockHash(height uint32, blockHash common.Uint256) {
	key := this.getBlockHashKey(height)
	this.store.BatchPut(key, blockHash.ToArray())
}

// SaveTransaction persist transaction to store
func (this *BlockStore) SaveTransaction(tx *types.Transaction, height uint32) error {
	if this.enableCache {
		this.cache.AddTransaction(tx, height)
	}
	return this.putTransaction(tx, height)
}

func (this *BlockStore) putTransaction(tx *types.Transaction, height uint32) error {
	txHash := tx.Hash()
	key := this.getTransactionKey(txHash)
	value := bytes.NewBuffer(nil)
	if err := serialization.WriteUint32(value, height); err != nil {
		return err
	}
	if err := serialization.WriteBytes(value, tx.Raw); err != nil {
		return err
	}
	this.store.BatchPut(key, value.Bytes())
	return nil
}

// GetTransaction return transaction by transaction hash
func (this *BlockStore) GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error) {
	if this.enableCache {
		tx, height := this.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}
	return this.loadTransaction(txHash)
}

func (this *BlockStore) loadTransaction(txHash common.Uint256) (*types.Transaction, uint32, error) {
	key := this.getTransactionKey(txHash)

	var tx *types.Transaction
	var height uint32
	if this.enableCache {
		tx, height = this.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}

	value, err := this.store.Get(key)
	if err != nil {
		return nil, 0, err
	}
	source := common.NewZeroCopySource(value)
	var eof bool
	height, eof = source.NextUint32()
	if eof {
		return nil, 0, io.ErrUnexpectedEOF
	}
	tx = new(types.Transaction)
	err = tx.Deserialization(source)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction deserialize error %s", err)
	}
	return tx, height, nil
}

// IsContainTransaction return whether the transaction is in store
func (this *BlockStore) ContainTransaction(txHash common.Uint256) (bool, error) {
	key := this.getTransactionKey(txHash)

	if this.enableCache {
		if this.cache.ContainTransaction(txHash) {
			return true, nil
		}
	}
	_, err := this.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetVersion return the version of store
func (this *BlockStore) GetVersion() (byte, error) {
	key := this.getVersionKey()
	value, err := this.store.Get(key)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(value)
	return reader.ReadByte()
}

// SaveVersion persist version to store
func (this *BlockStore) SaveVersion(ver byte) error {
	key := this.getVersionKey()
	return this.store.Put(key, []byte{ver})
}

// ClearAll clear all the data of block store
func (this *BlockStore) ClearAll() error {
	this.NewBatch()
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		this.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return this.CommitTo()
}

// CommitTo commit the batch to store
func (this *BlockStore) CommitTo() error {
	return this.store.BatchCommit()
}

// Close block store
func (this *BlockStore) Close() error {
	return this.store.Close()
}

func (this *BlockStore) getTransactionKey(txHash common.Uint256) []byte {
	key := bytes.NewBuffer(nil)
	key.WriteByte(byte(scom.DATA_TRANSACTION))
	txHash.Serialize(key)
	return key.Bytes()
}

func (this *BlockStore) getHeaderKey(blockHash common.Uint256) []byte {
	data := blockHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.DATA_HEADER)
	copy(key[1:], data)
	return key
}

func (this *BlockStore) getBlockHashKey(height uint32) []byte {
	key := make([]byte, 5, 5)
	key[0] = byte(scom.DATA_BLOCK)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key
}

func (this *BlockStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (this *BlockStore) getBlockMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_BLOCK_MERKLE_TREE)}
}

func (this *BlockStore) getVersionKey() []byte {
	return []byte{byte(scom.SYS_VERSION)}
}
