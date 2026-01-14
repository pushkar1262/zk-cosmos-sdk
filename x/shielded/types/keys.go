// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "shielded"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore key prefixes
var (
	// KeyPrefixCommitment stores commitments in Merkle tree
	KeyPrefixCommitment = []byte{0x01}

	// KeyPrefixNullifier stores nullifiers (prevents double-spend)
	KeyPrefixNullifier = []byte{0x02}

	// KeyPrefixMerkleRoot stores Merkle root per pool
	KeyPrefixMerkleRoot = []byte{0x03}

	// KeyPrefixCommitmentIndex stores commitment index in Merkle tree
	KeyPrefixCommitmentIndex = []byte{0x04}

	// KeyPrefixVerificationKey stores Groth16 verification keys
	KeyPrefixVerificationKey = []byte{0x05}

	// KeyPrefixPool stores pool information
	KeyPrefixPool = []byte{0x06}

	// KeyPrefixWithdrawalVerificationKey stores withdrawal verification keys
	KeyPrefixWithdrawalVerificationKey = []byte{0x07}
)

// GetNullifierKey returns the key for a nullifier
func GetNullifierKey(poolId uint64, nullifier []byte) []byte {
	return append(append(KeyPrefixNullifier, sdk.Uint64ToBigEndian(poolId)...), nullifier...)
}

// GetCommitmentIndexKey returns the key for a commitment index
func GetCommitmentIndexKey(poolId uint64, commitment []byte) []byte {
	return append(append(KeyPrefixCommitmentIndex, sdk.Uint64ToBigEndian(poolId)...), commitment...)
}

// GetMerkleRootKey returns the key for a Merkle root
func GetMerkleRootKey(poolId uint64) []byte {
	return append(KeyPrefixMerkleRoot, sdk.Uint64ToBigEndian(poolId)...)
}

// GetVerificationKeyKey returns the key for a verification key
func GetVerificationKeyKey(merkleDepth uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, merkleDepth)
	return append(KeyPrefixVerificationKey, buf...)
}

// GetWithdrawalVerificationKeyKey returns the key for a withdrawal verification key
func GetWithdrawalVerificationKeyKey() []byte {
	return KeyPrefixWithdrawalVerificationKey
}
