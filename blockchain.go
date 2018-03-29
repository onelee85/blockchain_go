package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	bolt "github.com/coreos/bbolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

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

//挖矿
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

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
func NewBlockchain(address string) *Blockchain {
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

			cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
			//创建创世块
			genesis := NewGenesisBlock(cbtx)

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

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
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

// 找到包含未花费输出的交易
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	//未花费输出的交易
	var unspentTXs []Transaction
	//已花费的输出
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator() //迭代区块链

	for {
		block := bci.Next()
		//每个区块中的交易记录
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			//每笔交易中的输出
		Outputs:
			for outIdx, out := range tx.Vout {
				// 查该输出是否已经被包含在一个交易的输入中
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				//如果一个输出被一个地址锁定，并且这个地址恰好是我们要找的地址
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			//判断是否为
			if tx.IsCoinbase() == false {
				//每笔交易输入
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// 对所有的未花费交易进行迭代，并对它的值进行累加。
//当累加值大于或等于我们想要传送的值时，它就会停止并返回累加值
//返回的还有通过交易 ID 进行分组的输出索引。我们只需取出足够支付的钱就够了。
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	//未花费的输出ID
	unspentOutputs := make(map[string][]int)
	//未花费的交易
	unspentTXs := bc.FindUnspentTransactions(address)
	//累加值
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
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
					fmt.Printf("Data: %s\n", block.HashTransactions())
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
