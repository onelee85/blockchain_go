/**
借鉴了Hashcash ，一个最初用来防止垃圾邮件的工作量证明算法。它可以被分解为以下步骤：
1.取一些公开的数据（比如，如果是 email 的话，它可以是接收者的邮件地址；在比特币中，它是区块头）
2.给这个公开数据添加一个计数器。计数器默认从 0 开始
3.将 data(数据) 和 counter(计数器) 组合到一起，获得一个哈希
4.检查哈希是否符合一定的条件：
	如果符合条件，结束
	如果不符合，增加计数器，重复步骤 3-4
**/
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//难度系数 开头有多少个0。24指的是算出来的哈希前 24 位必须是 0，如果用 16 进制表示，就是前 6 位必须是 0
const targetBits = 24

//工作量证明
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	//左移 256 - targetBits 位。
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}
	return pow
}

//准备数据
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

const maxNonce = math.MaxInt64

/**
把目标想象为一个范围的上界：如果一个数（由哈希转换而来）比上界要小，那么是有效的，反之无效。
因为要求比上界要小，所以会导致有效数字并不会很多。
因此，也就需要通过一些困难的工作（一系列反复地计算），才能找到一个有效的数字。
故：hash1无效， hash2有效
hash1:	0fac49161af82ed938add1d8725835cc123a1a87b1b196488360e58d4bfb51e3
traget: 0000010000000000000000000000000000000000000000000000000000000000
hash2:	0000008b0f41ec78bab747864db66bcb9fb89920ee75f43fdaaeb5544f7f76ca
**/
func (pow *ProofOfWork) run() (int, []byte) {
	var hasInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce) //准备数据
		hash = sha256.Sum256(data)     //用 SHA-256 对数据进行哈希
		hasInt.SetBytes(hash[:])       //将哈希转换成一个大整数
		//将这个大整数与目标进行比较 hasInt比target小
		if hasInt.Cmp(pow.target) == -1 {
			fmt.Printf("Hash: %x\n", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:]

}

//对工作量证明进行验证
func (pow *ProofOfWork) Validate() bool {
	var hasInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hasInt.SetBytes(hash[:])
	vaild := hasInt.Cmp(pow.target) == -1
	return vaild
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
