// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// HasNullifier checks if a nullifier exists (prevents double-spend)
func (k Keeper) HasNullifier(ctx sdk.Context, poolId uint64, nullifier []byte) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetNullifierKey(poolId, nullifier)
	return store.Has(key)
}

// SetNullifier stores a nullifier (marks commitment as spent)
func (k Keeper) SetNullifier(ctx sdk.Context, poolId uint64, nullifier []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetNullifierKey(poolId, nullifier)
	store.Set(key, []byte{1})
}

// HasWithdrawalNullifier checks if a withdrawal nullifier exists
func (k Keeper) HasWithdrawalNullifier(ctx sdk.Context, poolId uint64, nullifier []byte) bool {
	store := ctx.KVStore(k.storeKey)
	key := append(types.GetNullifierKey(poolId, nullifier), []byte("withdrawal")...)
	return store.Has(key)
}

// SetWithdrawalNullifier stores a withdrawal nullifier
func (k Keeper) SetWithdrawalNullifier(ctx sdk.Context, poolId uint64, nullifier []byte) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.GetNullifierKey(poolId, nullifier), []byte("withdrawal")...)
	store.Set(key, []byte{1})
}

