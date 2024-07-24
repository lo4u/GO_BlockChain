package merkletree

import (
	"GOPreject/transaction"
	"bytes"
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	LeftNode  *MerkleNode
	RightNode *MerkleNode
	HashData  []byte
}

func createNode(left, right *MerkleNode) *MerkleNode {
	tempNode := MerkleNode{}

	catenateHash := append(left.HashData, right.HashData...)
	hash := sha256.Sum256(catenateHash)
	tempNode.HashData = hash[:]
	tempNode.LeftNode = left
	tempNode.RightNode = right
	return &tempNode
}

func createLeaf(hashData []byte) *MerkleNode {
	return &MerkleNode{
		LeftNode:  nil,
		RightNode: nil,
		HashData:  hashData,
	}
}

func CreateMerkleTree(txs []*transaction.Transaction) *MerkleTree {
	txsLen := len(txs)
	if txsLen%2 != 0 {
		txs = append(txs, txs[txsLen-1])
		txsLen++
	}

	nodePool := make([]*MerkleNode, 0, txsLen)

	for _, tx := range txs {
		nodePool = append(nodePool, createLeaf(tx.ID))
	}
	for len(nodePool) > 1 {
		poolLen := len(nodePool)
		tempNodePool := make([]*MerkleNode, 0, poolLen/2+1)

		if poolLen%2 != 0 {
			tempNodePool = append(tempNodePool, nodePool[poolLen-1])
		}
		for i := 0; i < poolLen/2; i++ {
			tempNodePool = append(tempNodePool, createNode(nodePool[2*i], nodePool[2*i+1]))
		}
		nodePool = tempNodePool
	}
	return &MerkleTree{nodePool[0]}
}

// simple payment verification
func SPV(txid, mtRootHash []byte, route []int, hashRoute [][]byte) bool {
	routeLen := len(route)
	tempHash := txid

	for i := routeLen - 1; i >= 0; i-- {
		if route[i] == 0 {
			catenateHash := append(tempHash, hashRoute[i]...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else {
			catenateHash := append(hashRoute[i], tempHash...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		}
	}
	return bytes.Equal(tempHash, mtRootHash)
}

// A recursive function called by next function.
func (merkleNode *MerkleNode) Find(data []byte, route []int, hashRoute [][]byte) (bool, []int, [][]byte) {
	findFlag := false

	if bytes.Equal(merkleNode.HashData, data) {
		findFlag = true
		return findFlag, route, hashRoute
	} else {
		if merkleNode.LeftNode != nil {
			findFlag, tempRoute, tempHashRoute := merkleNode.LeftNode.Find(
				data,
				append(route, 0),
				append(hashRoute, merkleNode.RightNode.HashData),
			)
			if findFlag {
				return findFlag, tempRoute, tempHashRoute
			}
		}
		if merkleNode.RightNode != nil {
			findFlag, tempRoute, tempHashRoute := merkleNode.RightNode.Find(
				data,
				append(route, 1),
				append(hashRoute, merkleNode.LeftNode.HashData),
			)
			if findFlag {
				return findFlag, tempRoute, tempHashRoute
			}
		}
		return findFlag, route, hashRoute
	}
}

// Search for a hash data. Return a boolean value, a int slice navigate through the tree,
// and a hash's slice([ ][ ]byte) storing the SPV path.
func (merkleTree *MerkleTree) Find(hashData []byte) (bool, []int, [][]byte) {
	var route []int
	var hashRoute [][]byte
	return merkleTree.RootNode.Find(hashData, route, hashRoute)
}
