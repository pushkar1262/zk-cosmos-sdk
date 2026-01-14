// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/x/shielded/keeper"
	"github.com/cosmos/cosmos-sdk/x/shielded/types"
)

func TestKeeper(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	// Create mock keepers
	accountKeeper := &mockAccountKeeper{}
	bankKeeper := &mockBankKeeper{}

	k := keeper.NewKeeper(
		key,
		cdc,
		authority,
		accountKeeper,
		bankKeeper,
	)

	require.NotNil(t, k)
	require.Equal(t, authority, k.GetAuthority())
}

// Mock keepers for testing
type mockAccountKeeper struct{}

func (m *mockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *mockAccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return authtypes.NewBaseAccountWithAddress(addr)
}

func (m *mockAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {}

type mockBankKeeper struct{}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.ZeroInt())
}

