package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64  //当前时间戳，也就是区块创建的时间
	Data          []byte //区块存储的实际有效信息，也就是交易
	PrevBlockHash []byte //前一个块的哈希，即父哈希
	Hash          []byte //当前块的哈希
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	//简单的取了Block 结构的部分字段（Timestamp, Data 和 PrevBlockHash），
	//并将它们相互拼接起来，然后在拼接后的结果上计算一个 SHA-256
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

//创建新区块
func NewBlock(data string, PrevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), PrevBlockHash, []byte{}}
	block.SetHash()
	return block
}
