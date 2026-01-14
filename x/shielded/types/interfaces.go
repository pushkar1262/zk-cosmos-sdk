// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper interface
// Matches authkeeper.AccountKeeper interface
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected bank keeper interface
// Matches bankkeeper.Keeper interface
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// Note: These interfaces match the actual keeper interfaces from Cosmos SDK
// AccountKeeper matches authkeeper.AccountKeeper
// BankKeeper matches bankkeeper.Keeper (which is an interface)

