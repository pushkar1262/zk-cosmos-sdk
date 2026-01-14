package main

import (
	"bytes"

	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"

	"github.com/cosmos/cosmos-sdk/x/shielded/circuit"
	"github.com/cosmos/cosmos-sdk/x/shielded/zk"
)

func main() {
	setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
	
	proveDepositCmd := flag.NewFlagSet("prove-deposit", flag.ExitOnError)
	amount := proveDepositCmd.Uint64("amount", 0, "Amount to deposit")
	depositSecret := proveDepositCmd.String("secret", "", "Secret (hex or int)")
	depositSalt := proveDepositCmd.String("salt", "", "Salt (hex or int), optional (random if empty)")



	idCmd := flag.NewFlagSet("identity", flag.ExitOnError)
	idSecret := idCmd.String("secret", "", "Identity secret (hex or int)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		setupCmd.Parse(os.Args[2:])
		must(runSetup())

	case "prove-deposit":
		proveDepositCmd.Parse(os.Args[2:])
		if *amount == 0 || *depositSecret == "" {
			fmt.Println("missing --amount or --secret")
			proveDepositCmd.Usage()
			os.Exit(1)
		}
		must(runProveDeposit(*amount, *depositSecret, *depositSalt))

	case "identity":
		idCmd.Parse(os.Args[2:])
		if *idSecret == "" {
			fmt.Println("missing --secret")
			idCmd.Usage()
			os.Exit(1)
		}
		must(runIdentity(*idSecret))

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`
zk-cli commands:
  setup
        Generates proving (pk.bin) and verifying (vk.bin) keys for DepositCircuit.
  
  prove-deposit --amount <uint> --secret <string> [--salt <string>]
        Generates a deposit proof and commitment.
        Prints the 'evmosd tx shielded deposit' command.

  identity --secret <string>
        Computes identity commitment from secret.
`)
}

func must(err error) {
	if err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}

func runSetup() error {
	fmt.Println("Compiling DepositCircuit...")
	ccs, err := circuit.CompileDepositCircuit()
	if err != nil {
		return fmt.Errorf("failed to compile circuit: %w", err)
	}

	fmt.Println("Running trusted setup (Groth16)...")
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return fmt.Errorf("failed to setup: %w", err)
	}

	// Save proving key
	pkFile, err := os.Create("deposit_pk.bin")
	if err != nil {
		return fmt.Errorf("failed to create deposit_pk.bin: %w", err)
	}
	defer pkFile.Close()

	if _, err := pk.WriteTo(pkFile); err != nil {
		return fmt.Errorf("failed to write proving key: %w", err)
	}
	fmt.Println("✓ Proving key saved to deposit_pk.bin")

	// Save verification key
	vkFile, err := os.Create("deposit_vk.bin")
	if err != nil {
		return fmt.Errorf("failed to create deposit_vk.bin: %w", err)
	}
	defer vkFile.Close()

	if _, err := vk.WriteTo(vkFile); err != nil {
		return fmt.Errorf("failed to write verification key: %w", err)
	}
	fmt.Println("✓ Verification key saved to deposit_vk.bin")

	return nil
}

func runProveDeposit(amount uint64, secretStr, saltStr string) error {
	// Parse inputs
	secret := zk.MustBig(secretStr)
	amt := new(big.Int).SetUint64(amount)
	
	var salt *big.Int
	if saltStr == "" {
		var err error
		salt, err = zk.GenerateRandomBigInt()
		if err != nil {
			return err
		}
		fmt.Printf("Generated random salt: %s\n", salt.String())
	} else {
		salt = zk.MustBig(saltStr)
	}

	// Compute commitment
	commitment := zk.ComputeDepositCommitment(secret, salt, amt)
	fmt.Printf("Commitment: 0x%s\n", hex.EncodeToString(commitment.Bytes()))

	// Load proving key
	pkFile, err := os.Open("deposit_pk.bin")
	if err != nil {
		return fmt.Errorf("failed to open deposit_pk.bin: %w (did you run 'setup'?)", err)
	}
	defer pkFile.Close()
	
	pk := groth16.NewProvingKey(ecc.BN254)
	if _, err := pk.ReadFrom(pkFile); err != nil {
		return fmt.Errorf("failed to read proving key: %w", err)
	}

	// Compile constraint system (needed for witness)
	ccs, err := circuit.CompileDepositCircuit()
	if err != nil {
		return err
	}

	// Create witness
	witness, err := frontend.NewWitness(
		&circuit.DepositCircuit{
			Commitment: commitment,
			Amount:     amt,
			Secret:     secret,
			Salt:       salt,
		},
		ecc.BN254.ScalarField(),
	)
	if err != nil {
		return fmt.Errorf("failed to create witness: %w", err)
	}

	// Generate proof
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		return fmt.Errorf("failed to prove: %w", err)
	}

	var buf bytes.Buffer
	proof.WriteTo(&buf)
	proofBytes := buf.Bytes()
	proofHex := hex.EncodeToString(proofBytes)

	fmt.Println("\n📦 Proof (hex):")
	fmt.Println(proofHex)

	fmt.Println("\n🚀 Command to run:")
	// evmosd tx shielded deposit [pool-id] [amount] [commitment-hex] [proof-hex]
	fmt.Printf("evmosd tx shielded deposit 1 %d %s %s --from <key> --fees 5000aevmos --keyring-backend test --chain-id evmos_9001-1 -y\n", 
		amount, 
		hex.EncodeToString(commitment.Bytes()), 
		proofHex,
	)

	return nil
}

func runIdentity(secretStr string) error {
	secret := zk.MustBig(secretStr)
	comm, err := zk.ComputeIdentityCommitment(secret)
	if err != nil {
		return err
	}
	fmt.Printf("Identity Commitment: 0x%s\n", hex.EncodeToString(comm.Bytes()))
	return nil
}
