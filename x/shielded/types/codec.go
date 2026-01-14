// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global shielded module codec
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino JSON compatible msgs
	AminoCdc = codec.NewAminoCodec(amino) //nolint:staticcheck
)

const (
	// Amino names
	depositToShieldedName = "evmos/shielded/MsgDepositToShielded"
	privateSendName       = "evmos/shielded/MsgPrivateSend"
	withdrawName          = "evmos/shielded/MsgWithdrawFromCommitment"
)

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgDepositToShielded{},
		&MsgPrivateSend{},
		&MsgWithdrawFromCommitment{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/shielded interfaces and
// concrete types on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDepositToShielded{}, depositToShieldedName, nil)
	cdc.RegisterConcrete(&MsgPrivateSend{}, privateSendName, nil)
	cdc.RegisterConcrete(&MsgWithdrawFromCommitment{}, withdrawName, nil)
}

