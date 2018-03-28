package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64  //当前时间戳，也就是区块创建的时间
	Data          []byte //区块存储的实际有效信息，也就是交易
	PrevBlockHash []byte //前一个块的哈希，即父哈希
	Hash          []byte //当前块的哈希
	Nonce         int
}

//创建新区块
func NewBlock(data string, PrevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), PrevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block) //工作量证明
	nonce, hash := pow.run()     //开始挖矿
	block.Hash = hash
	block.Nonce = nonce
	return block
}

// 序列化, 由于BoltDB中值只能是 []byte 类型，但是我们想要存储 Block 结构
func (b *Block) Serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {

		log.Panic(err)
	}

	return &block
}
