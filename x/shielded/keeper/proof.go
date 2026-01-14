// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"io"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// VerifyDepositProof verifies the Groth16 proof for a deposit transaction
// The proof proves that commitment = MiMC(secret, salt, amount)
func (k Keeper) VerifyDepositProof(ctx sdk.Context, proofBytes []byte, msg *types.MsgDepositToShielded) error {
	if len(proofBytes) == 0 {
		return errorsmod.Wrap(types.ErrInvalidProof, "proof cannot be empty")
	}

	// Load verification key (depth 0 for deposit circuit)
	vkBytes, err := k.GetVerificationKey(ctx, 0)
	if err != nil {
		return errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to load deposit verification key")
	}

	// Deserialize verification key
	vk, err := deserializeVerificationKey(vkBytes)
	if err != nil {
		return errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to deserialize verification key")
	}

	// Deserialize proof
	proof, err := deserializeGroth16Proof(proofBytes)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidProof, "failed to deserialize proof")
	}

	// Prepare public inputs
	// For deposit: [commitment, amount]
	publicInputs := []*big.Int{
		new(big.Int).SetBytes(msg.Commitment),
		new(big.Int).SetUint64(msg.Amount),
	}

	// Create public witness
	publicWitness, err := createPublicWitness(publicInputs)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidProof, "failed to create public witness")
	}

	// Verify proof using Groth16
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		return errorsmod.Wrap(types.ErrProofVerificationFailed, err.Error())
	}

	return nil
}

// VerifyPrivateSendProof verifies the Groth16 proof for a private send transaction
// The proof proves all circuit constraints without revealing amounts
func (k Keeper) VerifyPrivateSendProof(ctx sdk.Context, proofBytes []byte, msg *types.MsgPrivateSend) error {
	if len(proofBytes) == 0 {
		return errorsmod.Wrap(types.ErrInvalidProof, "proof cannot be empty")
	}

	// Load verification key for this Merkle depth
	vkBytes, err := k.GetVerificationKey(ctx, msg.MerklePathSize)
	if err != nil {
		return errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to load verification key")
	}

	// Deserialize verification key
	vk, err := deserializeVerificationKey(vkBytes)
	if err != nil {
		return errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to deserialize verification key")
	}

	// Deserialize proof
	proof, err := deserializeGroth16Proof(proofBytes)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidProof, "failed to deserialize proof")
	}

	// Prepare public witness (must match circuit public inputs order)
	// Order: [nullifier, senderIdentity, recipientCommitment, changeCommitment, merkleRoot]
	publicInputs := []*big.Int{
		new(big.Int).SetBytes(msg.Nullifier),
		new(big.Int).SetBytes(msg.SenderIdentity),
		new(big.Int).SetBytes(msg.RecipientCommitment),
		new(big.Int).SetBytes(msg.ChangeCommitment),
		new(big.Int).SetBytes(msg.MerkleRoot),
	}

	// Create public witness
	publicWitness, err := createPublicWitness(publicInputs)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidProof, "failed to create public witness")
	}

	// Verify proof using Groth16
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		return errorsmod.Wrap(types.ErrProofVerificationFailed, err.Error())
	}

	return nil
}

// VerifyWithdrawalProof verifies the Groth16 proof for a withdrawal transaction
// The proof proves that commitment = MiMC(identity, salt, amount)
func (k Keeper) VerifyWithdrawalProof(ctx sdk.Context, proofBytes []byte, msg *types.MsgWithdrawFromCommitment) error {
	if len(proofBytes) == 0 {
		return errorsmod.Wrap(types.ErrInvalidProof, "proof cannot be empty")
	}

	// Load withdrawal verification key
	vkBytes, err := k.GetWithdrawalVerificationKey(ctx)
	if err != nil {
		return errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to load withdrawal verification key")
	}

	// Deserialize verification key
	vk, err := deserializeVerificationKey(vkBytes)
	if err != nil {
		return errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to deserialize verification key")
	}

	// Deserialize proof
	proof, err := deserializeGroth16Proof(proofBytes)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidProof, "failed to deserialize proof")
	}

	// Prepare public witness
	// Order: [commitment, recipientIdentity, withdrawalAmount, nullifier]
	publicInputs := []*big.Int{
		new(big.Int).SetBytes(msg.Commitment),
		new(big.Int).SetBytes(msg.RecipientIdentity),
		new(big.Int).SetUint64(msg.WithdrawalAmount),
		new(big.Int).SetBytes(msg.Nullifier),
	}

	// Create public witness
	publicWitness, err := createPublicWitness(publicInputs)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidProof, "failed to create public witness")
	}

	// Verify proof using Groth16
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidWithdrawalProof, err.Error())
	}

	return nil
}

// GetVerificationKey returns the verification key for a given Merkle depth
func (k Keeper) GetVerificationKey(ctx sdk.Context, merkleDepth uint32) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetVerificationKeyKey(merkleDepth)
	vk := store.Get(key)
	if vk == nil {
		return nil, errorsmod.Wrap(types.ErrVerificationKeyNotFound, "verification key not found")
	}
	return vk, nil
}

// SetVerificationKeyBytes stores a verification key for a given Merkle depth
func (k Keeper) SetVerificationKeyBytes(ctx sdk.Context, merkleDepth uint32, vk []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetVerificationKeyKey(merkleDepth)
	store.Set(key, vk)
}

// GetWithdrawalVerificationKey returns the withdrawal verification key
func (k Keeper) GetWithdrawalVerificationKey(ctx sdk.Context) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWithdrawalVerificationKeyKey()
	vk := store.Get(key)
	if vk == nil {
		return nil, errorsmod.Wrap(types.ErrVerificationKeyNotFound, "withdrawal verification key not found")
	}
	return vk, nil
}

// SetWithdrawalVerificationKey stores the withdrawal verification key
func (k Keeper) SetWithdrawalVerificationKey(ctx sdk.Context, vk []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWithdrawalVerificationKeyKey()
	store.Set(key, vk)
}

// deserializeGroth16Proof deserializes a Groth16 proof from bytes
// Groth16 proof format for BN254:
// - A (G1): 64 bytes (x: 32 bytes, y: 32 bytes)
// - B (G2): 128 bytes (x[0]: 32 bytes, x[1]: 32 bytes, y[0]: 32 bytes, y[1]: 32 bytes)
// - C (G1): 64 bytes (x: 32 bytes, y: 32 bytes)
// Total: 256 bytes
//
// gnark uses its own binary format for proof serialization
// This function reads from the gnark binary format
func deserializeGroth16Proof(proofBytes []byte) (groth16.Proof, error) {
	if len(proofBytes) == 0 {
		return nil, errorsmod.Wrap(types.ErrInvalidProof, "proof bytes empty")
	}

	// Create a new proof object for BN254 curve
	proof := groth16.NewProof(ecc.BN254)

	// Read the proof from bytes using gnark's ReaderFrom interface
	reader := &proofReader{data: proofBytes}
	if _, err := proof.ReadFrom(reader); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidProof, "failed to deserialize proof: "+err.Error())
	}

	return proof, nil
}

// proofReader is a helper to read proof data
type proofReader struct {
	data []byte
	pos  int
}

func (r *proofReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// deserializeVerificationKey deserializes a Groth16 verification key from bytes
// gnark uses its own binary format for verification key serialization
func deserializeVerificationKey(vkBytes []byte) (groth16.VerifyingKey, error) {
	if len(vkBytes) == 0 {
		return nil, errorsmod.Wrap(types.ErrVerificationKeyNotFound, "verification key bytes empty")
	}

	// Create a new verifying key for BN254 curve
	vk := groth16.NewVerifyingKey(ecc.BN254)

	// Read the verification key from bytes using gnark's ReaderFrom interface
	reader := &proofReader{data: vkBytes}
	if _, err := vk.ReadFrom(reader); err != nil {
		return nil, errorsmod.Wrap(types.ErrVerificationKeyNotFound, "failed to deserialize verification key: "+err.Error())
	}

	return vk, nil
}

// createPublicWitness creates a public witness from public inputs for Groth16 verification
func createPublicWitness(publicInputs []*big.Int) (witness.Witness, error) {
	// Create a witness for BN254 curve
	field := ecc.BN254.ScalarField()
	w, err := witness.New(field)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidProof, "failed to create witness")
	}

	// Create a channel to feed values to the witness
	// The witness expects: [public inputs..., secret inputs...]
	values := make(chan any, len(publicInputs))

	// Normalize and send public inputs
	for _, input := range publicInputs {
		fieldInput := new(big.Int).Mod(input, field)
		values <- fieldInput
	}
	close(values)

	// Fill the witness with public inputs (no secret inputs)
	if err := w.Fill(len(publicInputs), 0, values); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidProof, "failed to fill witness: "+err.Error())
	}

	return w, nil
}

// Helper function to convert bytes to big.Int (for proof verification)
func bytesToBigInt(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}
