// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
)

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Pools: []Pool{},
	}
}

// Validate performs basic genesis state validation
func (gs *GenesisState) Validate() error {
	for _, pool := range gs.Pools {
		if pool.PoolId == 0 {
			return fmt.Errorf("invalid pool id: %d", pool.PoolId)
		}
	}
	return nil
}

