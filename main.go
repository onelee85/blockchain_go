package main

func main() {
	//启动区块链
	blockchain := NewBlockchain()
	//关闭区块链
	defer blockchain.STOP()

	cli := CLI{blockchain}
	cli.Run()
}
