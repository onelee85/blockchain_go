package main

//区块链。本质上，区块链就是一个有着特定结构的数据库，是一个有序，每一个块都连接到前一个块的链表。
//也就是说，区块按照插入的顺序进行存储，每个块都与前一个块相连。
//这样的结构，能够让我们快速地获取链上的最新块，并且高效地通过哈希来检索一个块。
type Blockchain struct {
	Blockchains []*Block
}

func (bc *Blockchain) addBlock(data string) {
	//获取之前的区块
	pervBlock := bc.Blockchains[len(bc.Blockchains)-1]
	//生产新的区块
	block := NewBlock(data, pervBlock.Hash)
	//将最新的区块追加都链上
	bc.Blockchains = append(bc.Blockchains, block)
}

//创世块方法
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

//创建创世块
func NewBlockchain() *Blockchain {
	//go语言&表示获取存储的内存地址
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
