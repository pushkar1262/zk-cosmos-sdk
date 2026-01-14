// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"strconv"

	"cosmossdk.io/math"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

var _ types.MsgServer = &Keeper{}

// DepositToShielded handles initial deposit (amount revealed)
func (k Keeper) DepositToShielded(ctx context.Context, msg *types.MsgDepositToShielded) (*types.MsgDepositToShieldedResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// 1. Verify proof (proves commitment is valid)
	if err := k.VerifyDepositProof(sdkCtx, msg.Proof, msg); err != nil {
		return nil, err
	}

	// 2. Check if commitment already exists
	if k.HasCommitment(sdkCtx, msg.PoolId, msg.Commitment) {
		return nil, errorsmod.Wrap(types.ErrCommitmentExists, "commitment already exists")
	}

	// 3. Debit from sender's account (amount revealed here)
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	coins := sdk.NewCoins(sdk.NewCoin("utoken", math.NewIntFromUint64(msg.Amount)))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coins); err != nil {
		return nil, errorsmod.Wrap(err, "failed to debit from account")
	}

	// 4. Add commitment to Merkle tree
	if err := k.AddCommitmentToMerkleTree(sdkCtx, msg.PoolId, msg.Commitment); err != nil {
		return nil, err
	}

	// 5. Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute("sender", msg.Sender),
			sdk.NewAttribute("pool_id", strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute("amount", strconv.FormatUint(msg.Amount, 10)), // Revealed
			sdk.NewAttribute("commitment", hex.EncodeToString(msg.Commitment)),
		),
	)

	return &types.MsgDepositToShieldedResponse{
		Commitment: msg.Commitment,
	}, nil
}

// PrivateSend handles private transfer (amount hidden)
func (k Keeper) PrivateSend(ctx context.Context, msg *types.MsgPrivateSend) (*types.MsgPrivateSendResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// 1. Verify Merkle root matches current state
	storedRoot := k.GetMerkleRoot(sdkCtx, msg.PoolId)
	if !bytes.Equal(storedRoot, msg.MerkleRoot) {
		return nil, errorsmod.Wrap(types.ErrMerkleRootMismatch, "merkle root mismatch")
	}

	// 2. Check nullifier (prevent double-spend)
	if k.HasNullifier(sdkCtx, msg.PoolId, msg.Nullifier) {
		return nil, errorsmod.Wrap(types.ErrNullifierUsed, "nullifier already used")
	}

	// 3. Verify Groth16 proof (proves amount is valid without revealing it)
	if err := k.VerifyPrivateSendProof(sdkCtx, msg.Proof, msg); err != nil {
		return nil, err
	}

	// 4. Store nullifier
	k.SetNullifier(sdkCtx, msg.PoolId, msg.Nullifier)

	// 5. Add commitments to Merkle tree
	if err := k.AddCommitmentToMerkleTree(sdkCtx, msg.PoolId, msg.RecipientCommitment); err != nil {
		return nil, err
	}
	if err := k.AddCommitmentToMerkleTree(sdkCtx, msg.PoolId, msg.ChangeCommitment); err != nil {
		return nil, err
	}

	// 6. Emit event (NO AMOUNT!)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrivateSend,
			sdk.NewAttribute("sender", msg.Sender),
			sdk.NewAttribute("pool_id", strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute("nullifier", hex.EncodeToString(msg.Nullifier)),
			sdk.NewAttribute("recipient_commitment", hex.EncodeToString(msg.RecipientCommitment)),
			sdk.NewAttribute("change_commitment", hex.EncodeToString(msg.ChangeCommitment)),
			// NO AMOUNT ATTRIBUTE! ✅
		),
	)

	return &types.MsgPrivateSendResponse{
		RecipientCommitment: msg.RecipientCommitment,
		ChangeCommitment:    msg.ChangeCommitment,
	}, nil
}

// WithdrawFromCommitment handles withdrawal (amount revealed)
func (k Keeper) WithdrawFromCommitment(ctx context.Context, msg *types.MsgWithdrawFromCommitment) (*types.MsgWithdrawFromCommitmentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// 1. Verify commitment exists
	if !k.HasCommitment(sdkCtx, msg.PoolId, msg.Commitment) {
		return nil, errorsmod.Wrap(types.ErrCommitmentNotFound, "commitment not found")
	}

	// 2. Check if already withdrawn
	if k.HasWithdrawalNullifier(sdkCtx, msg.PoolId, msg.Nullifier) {
		return nil, errorsmod.Wrap(types.ErrCommitmentAlreadyWithdrawn, "commitment already withdrawn")
	}

	// 3. Verify withdrawal proof
	if err := k.VerifyWithdrawalProof(sdkCtx, msg.Proof, msg); err != nil {
		return nil, err
	}

	// 4. Mark commitment as withdrawn
	k.SetWithdrawalNullifier(sdkCtx, msg.PoolId, msg.Nullifier)
	k.MarkCommitmentWithdrawn(sdkCtx, msg.PoolId, msg.Commitment)

	// 5. Transfer tokens to receiver's account
	recipientAddr, err := sdk.AccAddressFromBech32(msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	coins := sdk.NewCoins(sdk.NewCoin("utoken", math.NewIntFromUint64(msg.WithdrawalAmount)))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coins); err != nil {
		return nil, errorsmod.Wrap(err, "failed to transfer tokens")
	}

	// 6. Emit event (amount revealed)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawal,
			sdk.NewAttribute("sender", msg.Sender),
			sdk.NewAttribute("recipient", msg.RecipientAddress),
			sdk.NewAttribute("pool_id", strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute("amount", strconv.FormatUint(msg.WithdrawalAmount, 10)), // Revealed
			sdk.NewAttribute("commitment", hex.EncodeToString(msg.Commitment)),
			sdk.NewAttribute("nullifier", hex.EncodeToString(msg.Nullifier)),
		),
	)

	return &types.MsgWithdrawFromCommitmentResponse{}, nil
}

// SetVerificationKey handles setting verification keys (via governance)
func (k Keeper) SetVerificationKey(ctx context.Context, msg *types.MsgSetVerificationKey) (*types.MsgSetVerificationKeyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Verify authority (must be governance module)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	if !authority.Equals(k.authority) {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"only governance module can set verification keys",
		)
	}

	// Store verification key based on circuit type
	switch msg.CircuitType {
	case types.CircuitTypeDeposit:
		// Deposit circuit uses depth 0
		k.SetVerificationKeyBytes(sdkCtx, 0, msg.VerificationKey)
		
	case types.CircuitTypePrivateSend:
		// Private send circuit uses merkle depth
		k.SetVerificationKeyBytes(sdkCtx, msg.MerkleDepth, msg.VerificationKey)
		
	case types.CircuitTypeWithdrawal:
		// Withdrawal circuit uses special key
		k.SetWithdrawalVerificationKey(sdkCtx, msg.VerificationKey)
		
	default:
		return nil, errorsmod.Wrap(types.ErrInvalidProof, "invalid circuit type")
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetVerificationKey,
			sdk.NewAttribute("circuit_type", msg.CircuitType),
			sdk.NewAttribute("merkle_depth", strconv.FormatUint(uint64(msg.MerkleDepth), 10)),
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgSetVerificationKeyResponse{}, nil
}

