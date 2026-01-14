package zk

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

const (
	MerkleDepth = 20
)

func mimcHash(left, right *big.Int) (*big.Int, error) {
	h := mimc.NewMiMC()
	// Write padded 32-bytes to match circuit behavior and avoid variable length issues (e.g. 0 -> nil)
	h.Write(ToBytes32(left))
	h.Write(ToBytes32(right))
	sum := h.Sum(nil)

	var out fr.Element
	out.SetBytes(sum)

	return out.BigInt(new(big.Int)), nil
}

// ToBytes32 ensures big.Int is converted to 32-byte big-endian slice
func ToBytes32(i *big.Int) []byte {
	b := i.Bytes()
	if len(b) == 32 {
		return b
	}
	
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

func stringToBigInt(v string) (*big.Int, error) {
	bi, ok := new(big.Int).SetString(v, 10)
	if !ok {
		return nil, errors.New("invalid big.Int string")
	}
	return bi, nil
}

func EmptyMerkleRoot() string {
	zero := big.NewInt(0)

	// zero leaf
	h := mimc.NewMiMC()
	h.Write(zero.Bytes())
	sum := h.Sum(nil)

	var node fr.Element
	node.SetBytes(sum)

	current := node.BigInt(new(big.Int))

	// build empty tree
	for i := 0; i < MerkleDepth; i++ {
		next, _ := mimcHash(current, current)
		current = next
	}

	return current.String()
}

func InsertLeaf(
	oldRoot string,
	index uint64,
	leaf string,
) (string, error) {

	current, err := stringToBigInt(leaf)
	if err != nil {
		return "", err
	}

	zero := big.NewInt(0)

	for level := 0; level < MerkleDepth; level++ {

		isRight := ((index >> level) & 1) == 1

		var left, right *big.Int

		if isRight {
			left = zero
			right = current
		} else {
			left = current
			right = zero
		}

		parent, err := mimcHash(left, right)
		if err != nil {
			return "", err
		}

		current = parent
	}

	return current.String(), nil
}

// BuildMerkleProof builds a Merkle proof for a given leaf index
// using MiMC hash (BN254), exactly matching on-chain InsertLeaf logic.
func BuildMerkleProof(
	leaves []string,
	targetIndex uint64,
) (
	root string,
	siblings []string,
	pathIndices []bool,
	err error,
) {

	if len(leaves) == 0 {
		return "", nil, nil, fmt.Errorf("empty merkle tree")
	}

	if targetIndex >= uint64(len(leaves)) {
		return "", nil, nil, fmt.Errorf("leaf index out of range")
	}

	// Parse leaves into big.Int
	nodes := make([]*big.Int, len(leaves))
	for i, l := range leaves {
		v, ok := new(big.Int).SetString(l, 10)
		if !ok {
			return "", nil, nil, fmt.Errorf("invalid leaf at index %d", i)
		}
		nodes[i] = v
	}

	index := targetIndex

	siblings = make([]string, 0, MerkleDepth)
	pathIndices = make([]bool, 0, MerkleDepth)

	// Build tree level by level
	for level := 0; level < MerkleDepth; level++ {

		// If odd number of nodes, pad with zero
		if len(nodes)%2 == 1 {
			nodes = append(nodes, big.NewInt(0))
		}

		nextLevel := make([]*big.Int, len(nodes)/2)

		for i := 0; i < len(nodes); i += 2 {

			left := nodes[i]
			right := nodes[i+1]

			parent, _ := mimcHash(left, right)
			nextLevel[i/2] = parent

			// Are we on the target path?
			if uint64(i) == index || uint64(i+1) == index {
				if index%2 == 0 {
					// target is left child
					siblings = append(siblings, right.String())
					pathIndices = append(pathIndices, false)
				} else {
					// target is right child
					siblings = append(siblings, left.String())
					pathIndices = append(pathIndices, true)
				}
				index = uint64(i / 2)
			}
		}

		nodes = nextLevel
	}

	if len(nodes) != 1 {
		return "", nil, nil, fmt.Errorf("invalid merkle construction")
	}

	root = nodes[0].String()

	return root, siblings, pathIndices, nil
}

// CalculateRoot computes the Merkle root for a set of leaves
func CalculateRoot(leaves []string) (string, error) {
	if len(leaves) == 0 {
		return EmptyMerkleRoot(), nil
	}

	// Parse leaves into big.Int
	nodes := make([]*big.Int, len(leaves))
	for i, l := range leaves {
		v, ok := new(big.Int).SetString(l, 10)
		if !ok {
			return "", fmt.Errorf("invalid leaf at index %d", i)
		}
		nodes[i] = v
	}

	// Build tree level by level
	for level := 0; level < MerkleDepth; level++ {
		// If odd number of nodes, pad with zero
		if len(nodes)%2 == 1 {
			nodes = append(nodes, big.NewInt(0))
		}

		nextLevel := make([]*big.Int, len(nodes)/2)

		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			right := nodes[i+1]
			parent, _ := mimcHash(left, right)
			nextLevel[i/2] = parent
		}

		nodes = nextLevel
	}

	if len(nodes) != 1 {
		return "", fmt.Errorf("invalid merkle construction")
	}

	return nodes[0].String(), nil
}
