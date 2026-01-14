package zk

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

// ComputeNullifier computes the nullifier = MiMC(secret, salt)
func ComputeNullifier(secret, salt *big.Int) *big.Int {
	s := new(big.Int).Mod(secret, fr.Modulus())
	sl := new(big.Int).Mod(salt, fr.Modulus())

	h := mimc.NewMiMC()
	h.Write(padTo32Bytes(s.Bytes()))
	h.Write(padTo32Bytes(sl.Bytes()))
	return new(big.Int).SetBytes(h.Sum(nil))
}

// ComputeIdentityCommitment computes the identity commitment = MiMC(secret)
func ComputeIdentityCommitment(secret *big.Int) (*big.Int, error) {
	s := new(big.Int).Mod(secret, fr.Modulus())

	h := mimc.NewMiMC()
	h.Write(padTo32Bytes(s.Bytes()))
	return new(big.Int).SetBytes(h.Sum(nil)), nil
}

// ComputeDepositCommitment computes commitment = MiMC(secret, salt, amount)
func ComputeDepositCommitment(secret, salt, amount *big.Int) *big.Int {
	s := new(big.Int).Mod(secret, fr.Modulus())
	sl := new(big.Int).Mod(salt, fr.Modulus())
	amt := new(big.Int).Mod(amount, fr.Modulus())

	h := mimc.NewMiMC()
	h.Write(padTo32Bytes(s.Bytes()))
	h.Write(padTo32Bytes(sl.Bytes()))
	h.Write(padTo32Bytes(amt.Bytes()))
	return new(big.Int).SetBytes(h.Sum(nil))
}

func padTo32Bytes(in []byte) []byte {
	out := make([]byte, 32)
	if len(in) > 32 {
		return in // Should not happen after Mod
	}
	copy(out[32-len(in):], in)
	return out
}

// Helper functions for CLI

func MustBig(s string) *big.Int {
	i, ok := new(big.Int).SetString(s, 0) // 0 detects base (hex, etc)
	if !ok {
		panic(fmt.Sprintf("invalid big int: %s", s))
	}
	// Reduce mod field to be safe
	i.Mod(i, fr.Modulus())
	return i
}

func ParseBigSlice(s string) []*big.Int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	res := make([]*big.Int, len(parts))
	for i, part := range parts {
		res[i] = MustBig(strings.TrimSpace(part))
	}
	return res
}

func ParseIndexSlice(s string) []*big.Int {
	// similar to ParseBigSlice but ensuring they are indices (0/1 usually for merkle)
	return ParseBigSlice(s)
}

func GenerateRandomBigInt() (*big.Int, error) {
	// Generate random 32 bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	i := new(big.Int).SetBytes(b)
	i.Mod(i, fr.Modulus()) // Ensure it's in the field
	return i, nil
}
