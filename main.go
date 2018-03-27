package main

import "fmt"

func main() {
	//创世块
	blockchain := NewBlockchain()
	//两笔交易记录
	blockchain.addBlock("send one doller to Bill")
	blockchain.addBlock("send 2 btcoins to James")

	for _, block := range blockchain.Blockchains {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
