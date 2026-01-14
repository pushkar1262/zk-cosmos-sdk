// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/shielded module sentinel errors
var (
	// Commitment errors
	ErrCommitmentNotFound    = errorsmod.Register(ModuleName, 1, "commitment not found")
	ErrCommitmentExists      = errorsmod.Register(ModuleName, 2, "commitment already exists")
	ErrInvalidCommitment     = errorsmod.Register(ModuleName, 3, "invalid commitment")
	
	// Nullifier errors
	ErrNullifierUsed         = errorsmod.Register(ModuleName, 10, "nullifier already used - double spend attempt")
	ErrInvalidNullifier      = errorsmod.Register(ModuleName, 11, "invalid nullifier")
	
	// Proof errors
	ErrInvalidProof          = errorsmod.Register(ModuleName, 20, "invalid zk-SNARK proof")
	ErrProofVerificationFailed = errorsmod.Register(ModuleName, 21, "proof verification failed")
	ErrVerificationKeyNotFound = errorsmod.Register(ModuleName, 22, "verification key not found")
	
	// Merkle tree errors
	ErrMerkleRootMismatch    = errorsmod.Register(ModuleName, 30, "merkle root mismatch")
	ErrInvalidMerkleProof    = errorsmod.Register(ModuleName, 31, "invalid merkle proof")
	ErrMerkleTreeFull        = errorsmod.Register(ModuleName, 32, "merkle tree is full")
	
	// Pool errors
	ErrPoolNotFound          = errorsmod.Register(ModuleName, 40, "pool not found")
	ErrInvalidPoolId         = errorsmod.Register(ModuleName, 41, "invalid pool id")
	
	// Amount errors
	ErrInsufficientBalance   = errorsmod.Register(ModuleName, 50, "insufficient balance")
	ErrInvalidAmount         = errorsmod.Register(ModuleName, 51, "invalid amount")
	ErrZeroAmount            = errorsmod.Register(ModuleName, 52, "amount must be positive")
	
	// Withdrawal errors
	ErrCommitmentAlreadyWithdrawn = errorsmod.Register(ModuleName, 60, "commitment already withdrawn")
	ErrInvalidWithdrawalProof = errorsmod.Register(ModuleName, 61, "invalid withdrawal proof")
)

