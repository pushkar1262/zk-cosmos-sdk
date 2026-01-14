// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package circuit

import (
	"bytes"
	"io"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// ============================================================================
// Key Generation and Management
// ============================================================================

// GenerateDepositKeys generates proving and verification keys for the deposit circuit
// This should be called during trusted setup
// Returns: proving key, verification key, and compiled constraint system
func GenerateDepositKeys() (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, error) {
	// Create a new circuit instance
	var circuit DepositCircuit

	// Compile the circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, nil, nil, err
	}

	// Run trusted setup (powers of tau)
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, nil, nil, err
	}

	return pk, vk, ccs, nil
}

// GeneratePrivateSendKeys generates proving and verification keys for the private send circuit
// This should be called during trusted setup for each Merkle depth
// merkleDepth: the depth of the Merkle tree (determines path length)
// Returns: proving key, verification key, and compiled constraint system
func GeneratePrivateSendKeys(merkleDepth uint32) (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, error) {
	// Create a new circuit instance with the specified Merkle depth
	circuit := PrivateTransferCircuit{
		MerklePath:    make([]frontend.Variable, merkleDepth),
		MerkleIndices: make([]frontend.Variable, merkleDepth),
	}

	// Compile the circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, nil, nil, err
	}

	// Run trusted setup (powers of tau)
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, nil, nil, err
	}

	return pk, vk, ccs, nil
}

// GenerateWithdrawalKeys generates proving and verification keys for the withdrawal circuit
// This should be called during trusted setup
// Returns: proving key, verification key, and compiled constraint system
func GenerateWithdrawalKeys() (groth16.ProvingKey, groth16.VerifyingKey, constraint.ConstraintSystem, error) {
	// Create a new circuit instance
	var circuit WithdrawalCircuit

	// Compile the circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, nil, nil, err
	}

	// Run trusted setup (powers of tau)
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, nil, nil, err
	}

	return pk, vk, ccs, nil
}

// ============================================================================
// Key Serialization
// ============================================================================

// SerializeProvingKey serializes a proving key to bytes
func SerializeProvingKey(pk groth16.ProvingKey) ([]byte, error) {
	var buf bytes.Buffer
	_, err := pk.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeProvingKey deserializes a proving key from bytes
func DeserializeProvingKey(keyBytes []byte) (groth16.ProvingKey, error) {
	pk := groth16.NewProvingKey(ecc.BN254)
	reader := bytes.NewReader(keyBytes)
	_, err := pk.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// SerializeVerifyingKey serializes a verification key to bytes
func SerializeVerifyingKey(vk groth16.VerifyingKey) ([]byte, error) {
	var buf bytes.Buffer
	_, err := vk.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeVerifyingKey deserializes a verification key from bytes
func DeserializeVerifyingKey(keyBytes []byte) (groth16.VerifyingKey, error) {
	vk := groth16.NewVerifyingKey(ecc.BN254)
	reader := bytes.NewReader(keyBytes)
	_, err := vk.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return vk, nil
}

// ============================================================================
// Key Storage Helpers
// ============================================================================

// SaveProvingKey saves a proving key to a file
func SaveProvingKey(pk groth16.ProvingKey, writer io.Writer) error {
	_, err := pk.WriteTo(writer)
	return err
}

// LoadProvingKey loads a proving key from a file
func LoadProvingKey(reader io.Reader) (groth16.ProvingKey, error) {
	pk := groth16.NewProvingKey(ecc.BN254)
	_, err := pk.ReadFrom(reader)
	return pk, err
}

// SaveVerifyingKey saves a verification key to a file
func SaveVerifyingKey(vk groth16.VerifyingKey, writer io.Writer) error {
	_, err := vk.WriteTo(writer)
	return err
}

// LoadVerifyingKey loads a verification key from a file
func LoadVerifyingKey(reader io.Reader) (groth16.VerifyingKey, error) {
	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err := vk.ReadFrom(reader)
	return vk, err
}
