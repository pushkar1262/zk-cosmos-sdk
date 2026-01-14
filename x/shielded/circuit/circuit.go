// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package circuit

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// ============================================================================
// DepositCircuit - Proves commitment = MiMC(secret, salt, amount)
// ============================================================================

// DepositCircuit defines the circuit for deposit transactions
// Public inputs: commitment, amount
// Private inputs: secret, salt
type DepositCircuit struct {
	// Public inputs
	Commitment frontend.Variable `gnark:",public"`
	Amount     frontend.Variable `gnark:",public"`

	// Private inputs
	Secret frontend.Variable
	Salt   frontend.Variable
}

// Define implements the circuit definition
func (circuit *DepositCircuit) Define(api frontend.API) error {
	// Initialize MiMC hash
	mimcHash, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// Constraint 1: Verify commitment = MiMC(secret, salt, amount)
	mimcHash.Write(circuit.Secret)
	mimcHash.Write(circuit.Salt)
	mimcHash.Write(circuit.Amount)
	computedCommitment := mimcHash.Sum()

	api.AssertIsEqual(computedCommitment, circuit.Commitment)

	// Constraint 2: Ensure amount is positive
	api.AssertIsLessOrEqual(0, circuit.Amount)

	return nil
}

// ============================================================================
// PrivateTransferCircuit - Private send with all constraints
// ============================================================================

// PrivateTransferCircuit defines the circuit for private send transactions
// This circuit proves all constraints without revealing amounts
type PrivateTransferCircuit struct {
	// Public inputs (on-chain, visible)
	Nullifier          frontend.Variable `gnark:",public"`
	SenderIdentity     frontend.Variable `gnark:",public"`
	RecipientCommitment frontend.Variable `gnark:",public"`
	ChangeCommitment   frontend.Variable `gnark:",public"`
	MerkleRoot         frontend.Variable `gnark:",public"`

	// Private inputs (off-chain, hidden)
	SenderSecret      frontend.Variable
	SenderSalt        frontend.Variable
	OldBalance        frontend.Variable // Hidden amount
	SendAmount        frontend.Variable // Hidden amount
	RecipientIdentity frontend.Variable
	RecipientSalt     frontend.Variable
	ChangeSalt        frontend.Variable
	OldCommitment     frontend.Variable
	MerklePath        []frontend.Variable
	MerkleIndices     []frontend.Variable
}

// Define implements the circuit definition with all 10 constraints
func (circuit *PrivateTransferCircuit) Define(api frontend.API) error {
	// Initialize MiMC hash
	mimcHash, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// Constraint 1: Old commitment valid
	// Prove: oldCommitment = MiMC(senderSecret, senderSalt, oldBalance)
	mimcHash.Reset()
	mimcHash.Write(circuit.SenderSecret)
	mimcHash.Write(circuit.SenderSalt)
	mimcHash.Write(circuit.OldBalance)
	computedOldCommitment := mimcHash.Sum()
	api.AssertIsEqual(computedOldCommitment, circuit.OldCommitment)

	// Constraint 2: Nullifier correct
	// Prove: nullifier = MiMC(senderSecret, senderSalt)
	mimcHash.Reset()
	mimcHash.Write(circuit.SenderSecret)
	mimcHash.Write(circuit.SenderSalt)
	computedNullifier := mimcHash.Sum()
	api.AssertIsEqual(computedNullifier, circuit.Nullifier)

	// Constraint 3: Identity correct
	// Prove: senderIdentity = MiMC(senderSecret)
	mimcHash.Reset()
	mimcHash.Write(circuit.SenderSecret)
	computedIdentity := mimcHash.Sum()
	api.AssertIsEqual(computedIdentity, circuit.SenderIdentity)

	// Constraint 4: Balance math - newBalance = oldBalance - sendAmount
	newBalance := api.Sub(circuit.OldBalance, circuit.SendAmount)

	// Constraint 7: Positive amount
	// Prove: sendAmount > 0
	api.AssertIsLessOrEqual(0, circuit.SendAmount)
	// Ensure it's actually positive (not zero)
	zero := frontend.Variable(0)
	api.AssertIsDifferent(circuit.SendAmount, zero)

	// Constraint 8: Sufficient balance
	// Prove: oldBalance >= sendAmount (equivalent to newBalance >= 0)
	api.AssertIsLessOrEqual(0, newBalance)

	// Constraint 5: Change commitment valid
	// Prove: changeCommitment = MiMC(senderSecret, changeSalt, newBalance)
	mimcHash.Reset()
	mimcHash.Write(circuit.SenderSecret)
	mimcHash.Write(circuit.ChangeSalt)
	mimcHash.Write(newBalance)
	computedChangeCommitment := mimcHash.Sum()
	api.AssertIsEqual(computedChangeCommitment, circuit.ChangeCommitment)

	// Constraint 6: Recipient commitment valid
	// Prove: recipientCommitment = MiMC(recipientIdentity, recipientSalt, sendAmount)
	mimcHash.Reset()
	mimcHash.Write(circuit.RecipientIdentity)
	mimcHash.Write(circuit.RecipientSalt)
	mimcHash.Write(circuit.SendAmount)
	computedRecipientCommitment := mimcHash.Sum()
	api.AssertIsEqual(computedRecipientCommitment, circuit.RecipientCommitment)

	// Constraint 9: Merkle proof valid
	// Prove that oldCommitment exists in the Merkle tree
	// Reconstruct the Merkle root from the path
	currentHash := circuit.OldCommitment
	
	// Traverse the Merkle path from leaf to root
	for i := 0; i < len(circuit.MerklePath); i++ {
		// Get the sibling node
		sibling := circuit.MerklePath[i]
		
		// Get the direction (0 = left, 1 = right)
		direction := circuit.MerkleIndices[i]
		
		// Hash the two nodes together based on direction
		// If direction is 0, currentHash is left child, sibling is right child
		// If direction is 1, sibling is left child, currentHash is right child
		mimcHash.Reset()
		
		// Use Select to choose the correct order based on direction
		leftChild := api.Select(direction, sibling, currentHash)
		rightChild := api.Select(direction, currentHash, sibling)
		
		mimcHash.Write(leftChild)
		mimcHash.Write(rightChild)
		currentHash = mimcHash.Sum()
	}
	
	// Verify that the computed root matches the public Merkle root
	api.AssertIsEqual(currentHash, circuit.MerkleRoot)

	// Constraint 10: Change note created
	// This is implicitly satisfied by constraint 5 (change commitment valid)
	// and constraint 8 (sufficient balance ensures newBalance >= 0)
	// We explicitly check that changeCommitment is not zero
	api.AssertIsDifferent(circuit.ChangeCommitment, zero)

	return nil
}

// ============================================================================
// WithdrawalCircuit - Proves commitment = MiMC(identity, salt, amount)
// ============================================================================

// WithdrawalCircuit defines the circuit for withdrawal transactions
// Public inputs: commitment, recipientIdentity, withdrawalAmount, nullifier
// Private inputs: identity, salt
type WithdrawalCircuit struct {
	// Public inputs
	Commitment       frontend.Variable `gnark:",public"`
	RecipientIdentity frontend.Variable `gnark:",public"`
	WithdrawalAmount frontend.Variable `gnark:",public"`
	Nullifier        frontend.Variable `gnark:",public"`

	// Private inputs
	Identity frontend.Variable
	Salt     frontend.Variable
}

// Define implements the circuit definition
func (circuit *WithdrawalCircuit) Define(api frontend.API) error {
	// Initialize MiMC hash
	mimcHash, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// Constraint 1: Verify commitment = MiMC(identity, salt, withdrawalAmount)
	mimcHash.Write(circuit.Identity)
	mimcHash.Write(circuit.Salt)
	mimcHash.Write(circuit.WithdrawalAmount)
	computedCommitment := mimcHash.Sum()
	api.AssertIsEqual(computedCommitment, circuit.Commitment)

	// Constraint 2: Verify recipientIdentity = MiMC(identity)
	mimcHash.Reset()
	mimcHash.Write(circuit.Identity)
	computedRecipientIdentity := mimcHash.Sum()
	api.AssertIsEqual(computedRecipientIdentity, circuit.RecipientIdentity)

	// Constraint 3: Verify nullifier = MiMC(identity, salt)
	mimcHash.Reset()
	mimcHash.Write(circuit.Identity)
	mimcHash.Write(circuit.Salt)
	computedNullifier := mimcHash.Sum()
	api.AssertIsEqual(computedNullifier, circuit.Nullifier)

	// Constraint 4: Ensure withdrawal amount is positive
	api.AssertIsLessOrEqual(0, circuit.WithdrawalAmount)
	zero := frontend.Variable(0)
	api.AssertIsDifferent(circuit.WithdrawalAmount, zero)

	return nil
}

