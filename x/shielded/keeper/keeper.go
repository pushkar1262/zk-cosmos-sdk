// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

// Keeper of this module maintains collections of shielded commitments
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this should be the x/gov module account.
	authority sdk.AccAddress
}

// NewKeeper creates new instances of the shielded Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority sdk.AccAddress,
	ak types.AccountKeeper,
	bk types.BankKeeper,
) Keeper {
	// ensure gov module account is set and is not nil
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		authority:     authority,
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() sdk.AccAddress {
	return k.authority
}

