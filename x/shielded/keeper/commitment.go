// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"bytes"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// HasCommitment checks if a commitment exists
func (k Keeper) HasCommitment(ctx sdk.Context, poolId uint64, commitment []byte) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCommitmentIndexKey(poolId, commitment)
	return store.Has(key)
}

// GetCommitmentIndex returns the index of a commitment in the Merkle tree
func (k Keeper) GetCommitmentIndex(ctx sdk.Context, poolId uint64, commitment []byte) int64 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCommitmentIndexKey(poolId, commitment)
	bz := store.Get(key)
	if bz == nil {
		return -1
	}
	// Convert uint64 to int64 (safe as index should be within int64 range)
	return int64(sdk.BigEndianToUint64(bz))
}

// SetCommitmentIndex sets the index of a commitment in the Merkle tree
func (k Keeper) SetCommitmentIndex(ctx sdk.Context, poolId uint64, commitment []byte, index uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCommitmentIndexKey(poolId, commitment)
	store.Set(key, sdk.Uint64ToBigEndian(index))
}

// MarkCommitmentWithdrawn marks a commitment as withdrawn
func (k Keeper) MarkCommitmentWithdrawn(ctx sdk.Context, poolId uint64, commitment []byte) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.GetCommitmentIndexKey(poolId, commitment), []byte("withdrawn")...)
	store.Set(key, []byte{1})
}

// IsCommitmentWithdrawn checks if a commitment has been withdrawn
func (k Keeper) IsCommitmentWithdrawn(ctx sdk.Context, poolId uint64, commitment []byte) bool {
	store := ctx.KVStore(k.storeKey)
	key := append(types.GetCommitmentIndexKey(poolId, commitment), []byte("withdrawn")...)
	return store.Has(key)
}

// GetAllCommitments returns all commitments for a pool
func (k Keeper) GetAllCommitments(ctx sdk.Context, poolId uint64) [][]byte {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixCommitmentIndex)
	prefixStore = prefix.NewStore(prefixStore, sdk.Uint64ToBigEndian(poolId))

	var commitments [][]byte
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Skip withdrawn commitments
		if bytes.HasSuffix(iterator.Key(), []byte("withdrawn")) {
			continue
		}
		commitments = append(commitments, iterator.Value())
	}

	return commitments
}

