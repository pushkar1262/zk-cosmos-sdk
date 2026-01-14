package keeper

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/shielded/types"
	"github.com/cosmos/cosmos-sdk/x/shielded/zk"
)

// GetMerkleRoot returns the Merkle root for a pool
func (k Keeper) GetMerkleRoot(ctx sdk.Context, poolId uint64) []byte {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMerkleRootKey(poolId)
	bz := store.Get(key)
	if bz == nil {
		// Return empty root (from zk)
		emptyRootStr := zk.EmptyMerkleRoot()
		// Convert to bytes
		root, _ := new(big.Int).SetString(emptyRootStr, 10)
		return zk.ToBytes32(root)
	}
	return bz
}

// SetMerkleRoot sets the Merkle root for a pool
func (k Keeper) SetMerkleRoot(ctx sdk.Context, poolId uint64, root []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMerkleRootKey(poolId)
	store.Set(key, root)
}

// AddCommitmentToMerkleTree adds a commitment to the Merkle tree and updates the root
func (k Keeper) AddCommitmentToMerkleTree(ctx sdk.Context, poolId uint64, commitment []byte) error {
	// Get all existing commitments
	commitments := k.GetAllCommitments(ctx, poolId)

	// Add new commitment
	commitments = append(commitments, commitment)

	// Store commitment index
	index := uint64(len(commitments) - 1)
	k.SetCommitmentIndex(ctx, poolId, commitment, index)

	// Transform [][]byte to []string (decimal) for MiMC
	leaves := make([]string, len(commitments))
	for i, c := range commitments {
		// Assume commitment is 32 bytes or less, interpret as big int
		v := new(big.Int).SetBytes(c)
		leaves[i] = v.String()
	}

	// Rebuild Merkle tree using ZK logic
	rootStr, err := zk.CalculateRoot(leaves)
	if err != nil {
		return err
	}

	// Convert root string to []byte
	rootInt, ok := new(big.Int).SetString(rootStr, 10)
	if !ok {
		return errors.New("invalid merkle root string")
	}
	rootBytes := zk.ToBytes32(rootInt)

	// Update root
	k.SetMerkleRoot(ctx, poolId, rootBytes)

	return nil
}

// GetMerkleProof generates a proof for a leaf
func (k Keeper) GetMerkleProof(ctx sdk.Context, poolId, leafIndex uint64) (root []byte, siblings [][]byte, pathIndices []bool, err error) {
	commitments := k.GetAllCommitments(ctx, poolId)
	if leafIndex >= uint64(len(commitments)) {
		return nil, nil, nil, errors.New("leaf index out of range")
	}

	leaves := make([]string, len(commitments))
	for i, c := range commitments {
		v := new(big.Int).SetBytes(c)
		leaves[i] = v.String()
	}

	rootStr, sibsStr, indices, err := zk.BuildMerkleProof(leaves, leafIndex)
	if err != nil {
		return nil, nil, nil, err
	}

	// Convert siblings strings to [][]byte
	siblings = make([][]byte, len(sibsStr))
	for i, s := range sibsStr {
		v, _ := new(big.Int).SetString(s, 10)
		siblings[i] = zk.ToBytes32(v)
	}

	rootInt, _ := new(big.Int).SetString(rootStr, 10)
	root = zk.ToBytes32(rootInt)

	return root, siblings, indices, nil
}
