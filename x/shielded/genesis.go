// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package shielded

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/keeper"
	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// InitGenesis initializes the shielded module's state from a genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Initialize pools
	for _, pool := range genState.Pools {
		if len(pool.MerkleRoot) > 0 {
			k.SetMerkleRoot(ctx, pool.PoolId, pool.MerkleRoot)
		}
	}

	// Set verification key from params if present
	if len(genState.Params.DepositVerificationKey) > 0 {
		msg := &types.MsgSetVerificationKey{
			Authority:       k.GetAuthority().String(),
			CircuitType:     "deposit",
			MerkleDepth:     0,
			VerificationKey: genState.Params.DepositVerificationKey,
		}
		// Directly call the keeper method which implements the logic, 
		// adhering to the signature found in previous error: (context.Context, *MsgSetVerificationKey)
		if _, err := k.SetVerificationKey(ctx, msg); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the shielded module's exported genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Export pools (simplified - in production, iterate through all pools)
	return types.DefaultGenesisState()
}

