package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

//subsidy 是挖出新块的奖励金。在比特币中，实际并没有存储这个数字，而是基于区块总数进行计算而得：区块总数除以 210000 就是 subsidy。
//挖出创世块的奖励是 50 BTC，每挖出 210000 个块后，奖励减半。在我们的实现中，这个奖励值将会是一个常量（至少目前是）。
const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput  //多笔交易输入
	Vout []TXOutput //多笔交易输出
}

//交易输出
type TXOutput struct {
	Value        int    //一定量的比特币
	ScriptPubKey string //一个锁定脚本(ScriptPubKey)，要花这笔钱，必须要解锁该脚本。
}

//交易输入
type TXInput struct {
	Txid      []byte
	Vout      int //每笔输入必须引用之前一笔交易的输出
	ScriptSig string
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

//创建一笔新的交易，将它放到一个块里，然后挖出这个块
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount) //找到所有的未花费输出，并且确保它们有足够的价值（value）

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

/**
coinbase 交易是一种特殊的交易，它不需要引用之前一笔交易的输出。
它“凭空”产生了币（也就是产生了新币），这是矿工获得挖出新块的奖励，也可以理解为“发行新币”。
**/
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
