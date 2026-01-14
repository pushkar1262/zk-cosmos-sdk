// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

var _ types.QueryServer = Keeper{}

// PublicBalance implements the Query/PublicBalance gRPC method.
func (k Keeper) PublicBalance(c context.Context, req *types.QueryPublicBalanceRequest) (*types.QueryPublicBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	// We use "utoken" as the default denom for this module
	balance := k.bankKeeper.GetBalance(sdkCtx, addr, "utoken")

	return &types.QueryPublicBalanceResponse{
		Balance: balance.Amount.Uint64(),
	}, nil
}

// PrivateBalance implements the Query/PrivateBalance gRPC method.
// Returns the total balance held in the shielded module account (pool budget).
func (k Keeper) PrivateBalance(c context.Context, req *types.QueryPrivateBalanceRequest) (*types.QueryPrivateBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	balance := k.bankKeeper.GetBalance(sdkCtx, moduleAddr, "utoken")

	return &types.QueryPrivateBalanceResponse{
		TotalBalance: balance.Amount.Uint64(),
	}, nil
}

// Commitments implements the Query/Commitments gRPC method.
func (k Keeper) Commitments(c context.Context, req *types.QueryCommitmentsRequest) (*types.QueryCommitmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	store := sdkCtx.KVStore(k.storeKey)
	commitmentStore := prefix.NewStore(store, append(types.KeyPrefixCommitmentIndex, sdk.Uint64ToBigEndian(req.PoolId)...))

	var commitments [][]byte
	pageRes, err := query.Paginate(commitmentStore, req.Pagination, func(key []byte, value []byte) error {
		// Value is the index, key is the commitment
		commitments = append(commitments, key)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCommitmentsResponse{
		Commitments: commitments,
		Pagination:  pageRes,
	}, nil
}

// MerkleRoot implements the Query/MerkleRoot gRPC method.
func (k Keeper) MerkleRoot(c context.Context, req *types.QueryMerkleRootRequest) (*types.QueryMerkleRootResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	root := k.GetMerkleRoot(sdkCtx, req.PoolId)

	return &types.QueryMerkleRootResponse{
		MerkleRoot: root,
	}, nil
}
