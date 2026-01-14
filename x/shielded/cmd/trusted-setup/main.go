// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

// trusted-setup is a command-line tool for generating Groth16 proving and verification keys
// for the shielded payment circuits. This should be run during the trusted setup ceremony.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/x/shielded/circuit"
)

func main() {
	var (
		circuitType    = flag.String("circuit", "", "Circuit type: deposit, private_send, or withdrawal")
		merkleDepth    = flag.Uint("merkle-depth", 0, "Merkle tree depth (only for private_send circuit)")
		outputDir      = flag.String("output", "./keys", "Output directory for keys")
		generateAll    = flag.Bool("all", false, "Generate keys for all circuits")
		maxMerkleDepth = flag.Uint("max-merkle-depth", 5, "Maximum Merkle depth for private_send (when using -all)")
	)

	flag.Parse()

	if *generateAll {
		generateAllKeys(*outputDir, *maxMerkleDepth)
		return
	}

	if *circuitType == "" {
		fmt.Fprintf(os.Stderr, "Error: circuit type is required (use -circuit flag or -all to generate all)\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	switch *circuitType {
	case "deposit":
		generateDepositKeys(*outputDir)
	case "private_send":
		if *merkleDepth == 0 {
			fmt.Fprintf(os.Stderr, "Error: merkle-depth must be > 0 for private_send circuit\n")
			os.Exit(1)
		}
		generatePrivateSendKeys(*outputDir, uint32(*merkleDepth))
	case "withdrawal":
		generateWithdrawalKeys(*outputDir)
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid circuit type: %s (must be deposit, private_send, or withdrawal)\n", *circuitType)
		os.Exit(1)
	}
}

func generateDepositKeys(outputDir string) {
	fmt.Println("Generating keys for deposit circuit...")

	pk, vk, ccs, err := circuit.GenerateDepositKeys()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating deposit keys: %v\n", err)
		os.Exit(1)
	}

	// Save proving key
	pkFile := filepath.Join(outputDir, "deposit_pk.bin")
	if err := saveProvingKey(pk, pkFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving proving key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Proving key saved to: %s\n", pkFile)

	// Save verification key
	vkFile := filepath.Join(outputDir, "deposit_vk.bin")
	if err := saveVerificationKey(vk, vkFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving verification key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Verification key saved to: %s\n", vkFile)

	// Save constraint system (for proof generation)
	ccsFile := filepath.Join(outputDir, "deposit_ccs.bin")
	if err := saveConstraintSystem(ccs, ccsFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving constraint system: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Constraint system saved to: %s\n", ccsFile)

	// Print verification key hash for verification
	vkHash := getKeyHash(vk)
	fmt.Printf("  Verification key hash: %s\n", hex.EncodeToString(vkHash))
	fmt.Println("✓ Deposit keys generated successfully")
}

func generatePrivateSendKeys(outputDir string, merkleDepth uint32) {
	fmt.Printf("Generating keys for private_send circuit (merkle depth: %d)...\n", merkleDepth)

	pk, vk, ccs, err := circuit.GeneratePrivateSendKeys(merkleDepth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating private_send keys: %v\n", err)
		os.Exit(1)
	}

	// Save proving key
	pkFile := filepath.Join(outputDir, fmt.Sprintf("private_send_depth_%d_pk.bin", merkleDepth))
	if err := saveProvingKey(pk, pkFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving proving key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Proving key saved to: %s\n", pkFile)

	// Save verification key
	vkFile := filepath.Join(outputDir, fmt.Sprintf("private_send_depth_%d_vk.bin", merkleDepth))
	if err := saveVerificationKey(vk, vkFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving verification key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Verification key saved to: %s\n", vkFile)

	// Save constraint system
	ccsFile := filepath.Join(outputDir, fmt.Sprintf("private_send_depth_%d_ccs.bin", merkleDepth))
	if err := saveConstraintSystem(ccs, ccsFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving constraint system: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Constraint system saved to: %s\n", ccsFile)

	// Print verification key hash
	vkHash := getKeyHash(vk)
	fmt.Printf("  Verification key hash: %s\n", hex.EncodeToString(vkHash))
	fmt.Printf("✓ Private send keys generated successfully (depth: %d)\n", merkleDepth)
}

func generateWithdrawalKeys(outputDir string) {
	fmt.Println("Generating keys for withdrawal circuit...")

	pk, vk, ccs, err := circuit.GenerateWithdrawalKeys()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating withdrawal keys: %v\n", err)
		os.Exit(1)
	}

	// Save proving key
	pkFile := filepath.Join(outputDir, "withdrawal_pk.bin")
	if err := saveProvingKey(pk, pkFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving proving key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Proving key saved to: %s\n", pkFile)

	// Save verification key
	vkFile := filepath.Join(outputDir, "withdrawal_vk.bin")
	if err := saveVerificationKey(vk, vkFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving verification key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Verification key saved to: %s\n", vkFile)

	// Save constraint system
	ccsFile := filepath.Join(outputDir, "withdrawal_ccs.bin")
	if err := saveConstraintSystem(ccs, ccsFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving constraint system: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Constraint system saved to: %s\n", ccsFile)

	// Print verification key hash
	vkHash := getKeyHash(vk)
	fmt.Printf("  Verification key hash: %s\n", hex.EncodeToString(vkHash))
	fmt.Println("✓ Withdrawal keys generated successfully")
}

func generateAllKeys(outputDir string, maxMerkleDepth uint) {
	fmt.Println("Generating keys for all circuits...")
	fmt.Println()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate deposit keys
	generateDepositKeys(outputDir)
	fmt.Println()

	// Generate withdrawal keys
	generateWithdrawalKeys(outputDir)
	fmt.Println()

	// Generate private_send keys for common Merkle depths
	fmt.Println("Generating private_send keys for common Merkle depths...")
	for depth := uint(1); depth <= maxMerkleDepth; depth++ {
		generatePrivateSendKeys(outputDir, uint32(depth))
		fmt.Println()
	}

	fmt.Println("✓ All keys generated successfully")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Verify all verification key hashes\n")
	fmt.Printf("2. Store verification keys on-chain via governance\n")
	fmt.Printf("3. Distribute proving keys securely to users\n")
}

func saveProvingKey(pk interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use WriteTo if available
	if writer, ok := pk.(interface {
		WriteTo(io.Writer) (int64, error)
	}); ok {
		_, err = writer.WriteTo(file)
		return err
	}

	return fmt.Errorf("proving key does not implement WriteTo")
}

func saveVerificationKey(vk interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use WriteTo if available
	if writer, ok := vk.(interface {
		WriteTo(io.Writer) (int64, error)
	}); ok {
		_, err = writer.WriteTo(file)
		return err
	}

	return fmt.Errorf("verification key does not implement WriteTo")
}

func saveConstraintSystem(ccs interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use WriteTo if available
	if writer, ok := ccs.(interface {
		WriteTo(io.Writer) (int64, error)
	}); ok {
		_, err = writer.WriteTo(file)
		return err
	}

	return fmt.Errorf("constraint system does not implement WriteTo")
}

func getKeyHash(key interface{}) []byte {
	// Simple hash of the key for verification purposes
	// In production, use a proper hash function
	if writer, ok := key.(interface {
		WriteTo(io.Writer) (int64, error)
	}); ok {
		var buf bytes.Buffer
		writer.WriteTo(&buf)
		hash := sha256.Sum256(buf.Bytes())
		return hash[:]
	}
	return nil
}
