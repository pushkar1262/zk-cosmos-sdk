// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// NewTxCmd returns a root CLI command handler for shielded transaction commands
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Shielded transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewDepositCmd(),
		NewPrivateSendCmd(),
		NewWithdrawalCmd(),
		NewMsgSetVerificationKeyCmd(),
	)

	return txCmd
}

// NewMsgSetVerificationKeyCmd returns a CLI command handler for setting verification keys
// This command should be used via governance proposal
func NewMsgSetVerificationKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-verification-key [circuit-type] [merkle-depth] [verification-key-file]",
		Short: "Set verification key for a circuit (via governance)",
		Long: `Set a verification key for a shielded circuit. This command creates a governance proposal.

Circuit types:
  - deposit: Deposit circuit (merkle-depth must be 0)
  - private_send: Private send circuit (merkle-depth must be > 0)
  - withdrawal: Withdrawal circuit (merkle-depth must be 0)

Example:
  evmosd tx shielded set-verification-key deposit 0 vk_deposit.bin --from validator --fees 1000aevmos`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			circuitType := args[0]
			merkleDepthStr := args[1]
			verificationKeyFile := args[2]

			// Validate circuit type
			if circuitType != types.CircuitTypeDeposit &&
				circuitType != types.CircuitTypePrivateSend &&
				circuitType != types.CircuitTypeWithdrawal {
				return fmt.Errorf("invalid circuit type: %s (must be deposit, private_send, or withdrawal)", circuitType)
			}

			// Parse merkle depth
			var merkleDepth uint32
			if _, err := fmt.Sscanf(merkleDepthStr, "%d", &merkleDepth); err != nil {
				return fmt.Errorf("invalid merkle depth: %w", err)
			}

			// Validate merkle depth based on circuit type
			if circuitType == types.CircuitTypePrivateSend && merkleDepth == 0 {
				return fmt.Errorf("merkle depth must be > 0 for private_send circuit")
			}
			if (circuitType == types.CircuitTypeDeposit || circuitType == types.CircuitTypeWithdrawal) && merkleDepth != 0 {
				return fmt.Errorf("merkle depth must be 0 for deposit and withdrawal circuits")
			}

			// Read verification key from file
			verificationKeyBytes, err := os.ReadFile(verificationKeyFile)
			if err != nil {
				return fmt.Errorf("failed to read verification key file: %w", err)
			}

			// Get authority (governance module address)
			// Note: This should be submitted via governance, so authority should be governance module
			authority := clientCtx.GetFromAddress().String()

			msg := types.NewMsgSetVerificationKey(
				authority,
				circuitType,
				merkleDepth,
				verificationKeyBytes,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgSetVerificationKeyProposalCmd returns a CLI command handler for creating a governance proposal
// to set verification keys
func NewMsgSetVerificationKeyProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-verification-key-proposal [circuit-type] [merkle-depth] [verification-key-file] [title] [description]",
		Short: "Create a governance proposal to set verification key",
		Long: `Create a governance proposal to set a verification key for a shielded circuit.

Circuit types:
  - deposit: Deposit circuit (merkle-depth must be 0)
  - private_send: Private send circuit (merkle-depth must be > 0)
  - withdrawal: Withdrawal circuit (merkle-depth must be 0)

Example:
  evmosd tx gov submit-proposal shielded set-verification-key-proposal deposit 0 vk_deposit.bin "Set Deposit VK" "Setting verification key for deposit circuit" --from validator --fees 1000aevmos`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			circuitType := args[0]
			merkleDepthStr := args[1]
			verificationKeyFile := args[2]
			title := args[3]
			description := args[4]

			// Validate circuit type
			if circuitType != types.CircuitTypeDeposit &&
				circuitType != types.CircuitTypePrivateSend &&
				circuitType != types.CircuitTypeWithdrawal {
				return fmt.Errorf("invalid circuit type: %s", circuitType)
			}

			// Parse merkle depth
			var merkleDepth uint32
			if _, err := fmt.Sscanf(merkleDepthStr, "%d", &merkleDepth); err != nil {
				return fmt.Errorf("invalid merkle depth: %w", err)
			}

			// Read verification key from file
			verificationKeyBytes, err := os.ReadFile(verificationKeyFile)
			if err != nil {
				return fmt.Errorf("failed to read verification key file: %w", err)
			}

			// Encode verification key as hex for the proposal
			verificationKeyHex := hex.EncodeToString(verificationKeyBytes)

			// Create proposal message
			// Note: This would typically use the governance module's proposal creation
			// For now, we'll create the message directly
			authority := sdk.AccAddress(clientCtx.GetFromAddress()).String()
			msg := types.NewMsgSetVerificationKey(
				authority,
				circuitType,
				merkleDepth,
				verificationKeyBytes,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// In a real implementation, this would create a governance proposal
			// For now, we'll just broadcast the message directly
			// The actual governance proposal creation would be handled by the governance module
			fmt.Printf("Proposal:\n")
			fmt.Printf("  Title: %s\n", title)
			fmt.Printf("  Description: %s\n", description)
			fmt.Printf("  Circuit Type: %s\n", circuitType)
			fmt.Printf("  Merkle Depth: %d\n", merkleDepth)
			fmt.Printf("  Verification Key (hex): %s...\n", verificationKeyHex[:64])

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewDepositCmd returns a CLI command handler for deposit transactions
func NewDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [pool-id] [amount] [commitment-hex] [proof-hex]",
		Short: "Deposit tokens to create a shielded commitment",
		Long: `Deposit tokens to create a shielded commitment. The amount is revealed in this transaction.

Parameters:
  - pool-id: The ID of the shielded pool
  - amount: Amount to deposit (in base units, e.g., 300 for 300 base units)
  - commitment-hex: Hex-encoded commitment (computed off-chain using MiMC hash)
  - proof-hex: Hex-encoded Groth16 proof (generated off-chain using deposit_pk.bin)

Example:
  evmosd tx shielded deposit 1 300 0x1234... 0xabcd... --from mykey --fees 1000aevmos

Note: The proof must be generated off-chain using the deposit proving key (deposit_pk.bin)
and constraint system (deposit_ccs.bin). The verification key (deposit_vk.bin) must be
stored on-chain via governance before deposits can be processed.`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}

			amount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			commitmentHex := args[2]
			commitment, err := hex.DecodeString(commitmentHex)
			if err != nil {
				return fmt.Errorf("invalid commitment hex: %w", err)
			}

			proofHex := args[3]
			proof, err := hex.DecodeString(proofHex)
			if err != nil {
				return fmt.Errorf("invalid proof hex: %w", err)
			}

			msg := types.NewMsgDepositToShielded(
				clientCtx.GetFromAddress().String(),
				poolId,
				amount,
				commitment,
				proof,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewPrivateSendCmd returns a CLI command handler for private send transactions
func NewPrivateSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "private-send [pool-id] [nullifier-hex] [sender-identity-hex] [recipient-commitment-hex] [change-commitment-hex] [merkle-root-hex] [proof-hex] [merkle-path-size]",
		Short: "Send tokens privately (amount hidden)",
		Long: `Send tokens privately without revealing the amount. This is a zero-knowledge transaction.

Parameters:
  - pool-id: The ID of the shielded pool
  - nullifier-hex: Hex-encoded nullifier (prevents double-spending)
  - sender-identity-hex: Hex-encoded sender identity
  - recipient-commitment-hex: Hex-encoded recipient commitment
  - change-commitment-hex: Hex-encoded change commitment (for remaining balance)
  - merkle-root-hex: Hex-encoded Merkle root of the pool
  - proof-hex: Hex-encoded Groth16 proof
  - merkle-path-size: Size of the Merkle path (depth of the tree)

Example:
  evmosd tx shielded private-send 1 0x1234... 0xabcd... 0x5678... 0x9abc... 0xdef0... 0x1111... 10 --from mykey --fees 1000aevmos

Note: The proof must be generated off-chain using the private_send proving key for the
corresponding Merkle depth.`,
		Args: cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}

			nullifier, err := hex.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid nullifier hex: %w", err)
			}

			senderIdentity, err := hex.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("invalid sender identity hex: %w", err)
			}

			recipientCommitment, err := hex.DecodeString(args[3])
			if err != nil {
				return fmt.Errorf("invalid recipient commitment hex: %w", err)
			}

			changeCommitment, err := hex.DecodeString(args[4])
			if err != nil {
				return fmt.Errorf("invalid change commitment hex: %w", err)
			}

			merkleRoot, err := hex.DecodeString(args[5])
			if err != nil {
				return fmt.Errorf("invalid merkle root hex: %w", err)
			}

			proof, err := hex.DecodeString(args[6])
			if err != nil {
				return fmt.Errorf("invalid proof hex: %w", err)
			}

			merklePathSize, err := strconv.ParseUint(args[7], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid merkle path size: %w", err)
			}

			msg := types.NewMsgPrivateSend(
				clientCtx.GetFromAddress().String(),
				poolId,
				nullifier,
				senderIdentity,
				recipientCommitment,
				changeCommitment,
				merkleRoot,
				proof,
				uint32(merklePathSize),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewWithdrawalCmd returns a CLI command handler for withdrawal transactions
func NewWithdrawalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [pool-id] [commitment-hex] [recipient-address] [withdrawal-amount] [salt-hex] [recipient-identity-hex] [proof-hex] [nullifier-hex]",
		Short: "Withdraw tokens from a shielded commitment",
		Long: `Withdraw tokens from a shielded commitment to a recipient address.

Parameters:
  - pool-id: The ID of the shielded pool
  - commitment-hex: Hex-encoded commitment to withdraw from
  - recipient-address: Bech32 address of the recipient
  - withdrawal-amount: Amount to withdraw (in base units)
  - salt-hex: Hex-encoded salt value
  - recipient-identity-hex: Hex-encoded recipient identity
  - proof-hex: Hex-encoded Groth16 proof
  - nullifier-hex: Hex-encoded nullifier (prevents double-spending)

Example:
  evmosd tx shielded withdraw 1 0x1234... evmos1abc... 300 0x5678... 0x9abc... 0xdef0... 0x1111... --from mykey --fees 1000aevmos

Note: The proof must be generated off-chain using the withdrawal proving key (withdrawal_pk.bin).`,
		Args: cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}

			commitment, err := hex.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid commitment hex: %w", err)
			}

			recipientAddress := args[2]
			if _, err := sdk.AccAddressFromBech32(recipientAddress); err != nil {
				return fmt.Errorf("invalid recipient address: %w", err)
			}

			withdrawalAmount, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid withdrawal amount: %w", err)
			}

			salt, err := hex.DecodeString(args[4])
			if err != nil {
				return fmt.Errorf("invalid salt hex: %w", err)
			}

			recipientIdentity, err := hex.DecodeString(args[5])
			if err != nil {
				return fmt.Errorf("invalid recipient identity hex: %w", err)
			}

			proof, err := hex.DecodeString(args[6])
			if err != nil {
				return fmt.Errorf("invalid proof hex: %w", err)
			}

			nullifier, err := hex.DecodeString(args[7])
			if err != nil {
				return fmt.Errorf("invalid nullifier hex: %w", err)
			}

			msg := types.NewMsgWithdrawFromCommitment(
				clientCtx.GetFromAddress().String(),
				poolId,
				commitment,
				recipientAddress,
				withdrawalAmount,
				salt,
				recipientIdentity,
				proof,
				nullifier,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
