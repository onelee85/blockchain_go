package main

import (
	"fmt"
	"log"
	"strconv"

	bolt "github.com/coreos/bbolt"
)

//区块链。本质上，区块链就是一个有着特定结构的数据库，是一个有序，每一个块都连接到前一个块的链表。
//也就是说，区块按照插入的顺序进行存储，每个块都与前一个块相连。
//这样的结构，能够让我们快速地获取链上的最新块，并且高效地通过哈希来检索一个块。
//仅存储区块链的 tip。另外，我们存储了一个数据库连接
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//我们会从数据库中获取最后一个块的哈希
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	//生产新的哈希块
	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash
		return nil
	})
}

//创世块方法
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

//go get github.com/coreos/bbolt/...
/**
1.打开一个数据库文件
2.检查文件里面是否已经存储了一个区块链
3.如果已经存储了一个区块链：
	创建一个新的 Blockchain 实例
	设置 Blockchain 实例的 tip 为数据库中存储的最后一个块的哈希
4.如果没有区块链：
	创建创世块
	存储到数据库
	将创世块哈希保存为最后一个块的哈希
	创建一个新的 Blockchain 实例，初始时 tip 指向创世块（tip 有尾部，尖端的意思，在这里 tip 存储的是最后一个块的哈希）
**/
func NewBlockchain() *Blockchain {
	var tip []byte
	//打开一个数据库文件
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//检查文件里面是否已经存储了一个区块链
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil { //如果没有区块链
			fmt.Println("No existing blockchain found. Creating a new one...")
			//创建创世块
			genesis := NewGenesisBlock()

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			//存储到数据库
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//设置 Blockchain 实例的 tip 为数据库中存储的最后一个块的哈希
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else { //如果已经存储了一个区块链：
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	//创建一个新的 Blockchain 实例
	//设置 Blockchain 实例的 tip 为数据库中存储的最后一个块的哈希
	bc := Blockchain{tip, db}

	return &bc
}

//关闭数据库
func (bc *Blockchain) STOP() {
	bc.db.Close()
}

// Iterator ...
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

//打印
func (bc *Blockchain) PrintAll() {
	bc.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(blocksBucket))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				if len(k) == 32 {
					block := DeserializeBlock(v)
					fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
					fmt.Printf("Data: %s\n", block.Data)
					fmt.Printf("Hash: %x\n", block.Hash)
					pow := NewProofOfWork(block)
					fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
					fmt.Println()
				}
				return nil
			})
		}

		return nil
	})
}
