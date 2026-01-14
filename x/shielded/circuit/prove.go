// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package circuit

import (
	"bytes"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/constraint"
)

// ============================================================================
// Deposit Proof Generation
// ============================================================================

// DepositProofInput contains all inputs needed to generate a deposit proof
type DepositProofInput struct {
	Secret     *big.Int
	Salt       *big.Int
	Amount     *big.Int
	Commitment *big.Int
}

// GenerateDepositProof generates a Groth16 proof for a deposit transaction
// This should be called off-chain by the user's wallet
// ccs: compiled constraint system (should be compiled once and reused)
// pk: proving key (from trusted setup)
func GenerateDepositProof(
	ccs constraint.ConstraintSystem,
	pk groth16.ProvingKey,
	input *DepositProofInput,
) (groth16.Proof, error) {
	// Create circuit assignment
	assignment := DepositCircuit{
		Commitment: input.Commitment,
		Amount:     input.Amount,
		Secret:     input.Secret,
		Salt:       input.Salt,
	}

	// Create witness from assignment
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	// Generate proof
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// ============================================================================
// Private Send Proof Generation
// ============================================================================

// PrivateSendProofInput contains all inputs needed to generate a private send proof
type PrivateSendProofInput struct {
	// Public inputs
	Nullifier          *big.Int
	SenderIdentity     *big.Int
	RecipientCommitment *big.Int
	ChangeCommitment   *big.Int
	MerkleRoot         *big.Int

	// Private inputs
	SenderSecret      *big.Int
	SenderSalt        *big.Int
	OldBalance        *big.Int
	SendAmount        *big.Int
	RecipientIdentity *big.Int
	RecipientSalt     *big.Int
	ChangeSalt        *big.Int
	OldCommitment     *big.Int
	MerklePath        []*big.Int
	MerkleIndices     []*big.Int
}

// GeneratePrivateSendProof generates a Groth16 proof for a private send transaction
// This should be called off-chain by the user's wallet
// ccs: compiled constraint system (should be compiled once and reused)
// pk: proving key (from trusted setup)
func GeneratePrivateSendProof(
	ccs constraint.ConstraintSystem,
	pk groth16.ProvingKey,
	input *PrivateSendProofInput,
) (groth16.Proof, error) {
	// Convert Merkle path and indices to frontend.Variable
	merklePathVars := make([]frontend.Variable, len(input.MerklePath))
	for i, v := range input.MerklePath {
		merklePathVars[i] = v
	}

	merkleIndicesVars := make([]frontend.Variable, len(input.MerkleIndices))
	for i, v := range input.MerkleIndices {
		merkleIndicesVars[i] = v
	}

	// Create circuit assignment
	assignment := PrivateTransferCircuit{
		Nullifier:          input.Nullifier,
		SenderIdentity:     input.SenderIdentity,
		RecipientCommitment: input.RecipientCommitment,
		ChangeCommitment:   input.ChangeCommitment,
		MerkleRoot:         input.MerkleRoot,
		SenderSecret:       input.SenderSecret,
		SenderSalt:         input.SenderSalt,
		OldBalance:         input.OldBalance,
		SendAmount:         input.SendAmount,
		RecipientIdentity:  input.RecipientIdentity,
		RecipientSalt:      input.RecipientSalt,
		ChangeSalt:         input.ChangeSalt,
		OldCommitment:      input.OldCommitment,
		MerklePath:         merklePathVars,
		MerkleIndices:      merkleIndicesVars,
	}

	// Create witness from assignment
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	// Generate proof
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// ============================================================================
// Withdrawal Proof Generation
// ============================================================================

// WithdrawalProofInput contains all inputs needed to generate a withdrawal proof
type WithdrawalProofInput struct {
	// Public inputs
	Commitment        *big.Int
	RecipientIdentity *big.Int
	WithdrawalAmount  *big.Int
	Nullifier         *big.Int

	// Private inputs
	Identity *big.Int
	Salt     *big.Int
}

// GenerateWithdrawalProof generates a Groth16 proof for a withdrawal transaction
// This should be called off-chain by the user's wallet
// ccs: compiled constraint system (should be compiled once and reused)
// pk: proving key (from trusted setup)
func GenerateWithdrawalProof(
	ccs constraint.ConstraintSystem,
	pk groth16.ProvingKey,
	input *WithdrawalProofInput,
) (groth16.Proof, error) {
	// Create circuit assignment
	assignment := WithdrawalCircuit{
		Commitment:       input.Commitment,
		RecipientIdentity: input.RecipientIdentity,
		WithdrawalAmount: input.WithdrawalAmount,
		Nullifier:        input.Nullifier,
		Identity:         input.Identity,
		Salt:             input.Salt,
	}

	// Create witness from assignment
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	// Generate proof
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// ============================================================================
// Circuit Compilation Helpers
// ============================================================================

// CompileDepositCircuit compiles the deposit circuit and returns the constraint system
// This should be called once and the result reused for proof generation
func CompileDepositCircuit() (constraint.ConstraintSystem, error) {
	var circuit DepositCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	return ccs, err
}

// CompilePrivateSendCircuit compiles the private send circuit and returns the constraint system
// merkleDepth: the depth of the Merkle tree
// This should be called once per Merkle depth and the result reused
func CompilePrivateSendCircuit(merkleDepth uint32) (constraint.ConstraintSystem, error) {
	circuit := PrivateTransferCircuit{
		MerklePath:    make([]frontend.Variable, merkleDepth),
		MerkleIndices: make([]frontend.Variable, merkleDepth),
	}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	return ccs, err
}

// CompileWithdrawalCircuit compiles the withdrawal circuit and returns the constraint system
// This should be called once and the result reused for proof generation
func CompileWithdrawalCircuit() (constraint.ConstraintSystem, error) {
	var circuit WithdrawalCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	return ccs, err
}

// ============================================================================
// Proof Serialization
// ============================================================================

// SerializeProof serializes a Groth16 proof to bytes for on-chain submission
func SerializeProof(proof groth16.Proof) ([]byte, error) {
	// Use gnark's WriterTo interface to serialize the proof
	var buf bytes.Buffer
	_, err := proof.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeProof deserializes a Groth16 proof from bytes
func DeserializeProof(proofBytes []byte) (groth16.Proof, error) {
	proof := groth16.NewProof(ecc.BN254)
	reader := bytes.NewReader(proofBytes)
	_, err := proof.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

