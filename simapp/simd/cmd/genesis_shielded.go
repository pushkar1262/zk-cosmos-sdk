package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// AddShieldedVkCmd returns a command that adds a shielded verification key to the genesis file
func AddShieldedVkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-shielded-vk [vk-file]",
		Short: "Add a shielded verification key to genesis.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			vkFile := args[0]
			vkBytes, err := os.ReadFile(vkFile)
			if err != nil {
				return fmt.Errorf("failed to read verification key file: %w", err)
			}

			genFile := config.GenesisFile()
			appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to read genesis file: %w", err)
			}

			var appState map[string]json.RawMessage
			if err := json.Unmarshal(appGenesis.AppState, &appState); err != nil {
				return fmt.Errorf("failed to unmarshal app state: %w", err)
			}

			var shieldedGenState types.GenesisState
			if appState[types.ModuleName] != nil {
				if err := clientCtx.Codec.UnmarshalJSON(appState[types.ModuleName], &shieldedGenState); err != nil {
					return fmt.Errorf("failed to unmarshal shielded genesis state: %w", err)
				}
			} else {
				// Default state if not present
				shieldedGenState = *types.DefaultGenesisState()
			}

			shieldedGenState.Params.DepositVerificationKey = vkBytes

			shieldedGenStateBz, err := clientCtx.Codec.MarshalJSON(&shieldedGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal shielded genesis state: %w", err)
			}

			appState[types.ModuleName] = shieldedGenStateBz
			appGenesis.AppState, err = json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal app state: %w", err)
			}

			if err := genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
				return fmt.Errorf("failed to export genesis file: %w", err)
			}

			fmt.Printf("Successfully added verification key from %s to %s\n", vkFile, genFile)
			return nil
		},
	}

	return cmd
}
