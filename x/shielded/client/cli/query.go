// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group shielded queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryPublicBalance(),
		GetCmdQueryPrivateBalance(),
		GetCmdQueryCommitments(),
		GetCmdQueryMerkleRoot(),
	)

	return cmd
}

// GetCmdQueryPublicBalance implements the public balance query command.
func GetCmdQueryPublicBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance-public [address]",
		Short: "Query the public balance of an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryPublicBalanceRequest{
				Address: args[0],
			}

			res, err := queryClient.PublicBalance(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("Public balance: %d utoken\n", res.Balance))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPrivateBalance implements the private balance query command.
func GetCmdQueryPrivateBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance-private [pool-id]",
		Short: "Query the total private balance in a pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryPrivateBalanceRequest{
				PoolId: poolId,
			}

			res, err := queryClient.PrivateBalance(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("Total private balance (pool %d): %d utoken\n", poolId, res.TotalBalance))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCommitments implements the commitments query command.
func GetCmdQueryCommitments() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commitments [pool-id]",
		Short: "Query all commitments in a pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryCommitmentsRequest{
				PoolId:     poolId,
				Pagination: pageReq,
			}

			res, err := queryClient.Commitments(cmd.Context(), params)
			if err != nil {
				return err
			}

			for _, commitment := range res.Commitments {
				fmt.Printf("- 0x%s\n", hex.EncodeToString(commitment))
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "commitments")

	return cmd
}

// GetCmdQueryMerkleRoot implements the Merkle root query command.
func GetCmdQueryMerkleRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merkle-root [pool-id]",
		Short: "Query the current Merkle root of a pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryMerkleRootRequest{
				PoolId: poolId,
			}

			res, err := queryClient.MerkleRoot(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("Merkle root: 0x%s\n", hex.EncodeToString(res.MerkleRoot)))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
