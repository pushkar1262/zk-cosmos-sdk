// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg              = &MsgDepositToShielded{}
	_ sdk.Msg              = &MsgPrivateSend{}
	_ sdk.Msg              = &MsgWithdrawFromCommitment{}
	_ sdk.Msg              = &MsgSetVerificationKey{}
	_ sdk.HasValidateBasic = &MsgDepositToShielded{}
	_ sdk.HasValidateBasic = &MsgPrivateSend{}
	_ sdk.HasValidateBasic = &MsgWithdrawFromCommitment{}
	_ sdk.HasValidateBasic = &MsgSetVerificationKey{}
)

const (
	TypeMsgDepositToShielded      = "deposit_to_shielded"
	TypeMsgPrivateSend            = "private_send"
	TypeMsgWithdrawFromCommitment = "withdraw_from_commitment"
	TypeMsgSetVerificationKey     = "set_verification_key"
	
	// Circuit types
	CircuitTypeDeposit     = "deposit"
	CircuitTypePrivateSend = "private_send"
	CircuitTypeWithdrawal  = "withdrawal"
)

// ============================================================================
// MsgDepositToShielded - Initial deposit (amount revealed)
// ============================================================================

// NewMsgDepositToShielded creates a new MsgDepositToShielded instance
func NewMsgDepositToShielded(
	sender string,
	poolId uint64,
	amount uint64,
	commitment []byte,
	proof []byte,
) *MsgDepositToShielded {
	return &MsgDepositToShielded{
		Sender:     sender,
		PoolId:     poolId,
		Amount:     amount,
		Commitment: commitment,
		Proof:     proof,
	}
}

// Route implements sdk.Msg
func (msg MsgDepositToShielded) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgDepositToShielded) Type() string { return TypeMsgDepositToShielded }

// ValidateBasic implements sdk.HasValidateBasic
func (msg MsgDepositToShielded) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrap(errortypes.ErrInvalidAddress, "invalid sender address")
	}
	if msg.PoolId == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "pool id cannot be zero")
	}
	if msg.Amount == 0 {
		return errorsmod.Wrap(ErrZeroAmount, "amount must be positive")
	}
	if len(msg.Commitment) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "commitment cannot be empty")
	}
	if len(msg.Proof) == 0 {
		return errorsmod.Wrap(ErrInvalidProof, "proof cannot be empty")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgDepositToShielded) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgDepositToShielded) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ============================================================================
// MsgPrivateSend - Private transfer (amount hidden)
// ============================================================================

// NewMsgPrivateSend creates a new MsgPrivateSend instance
func NewMsgPrivateSend(
	sender string,
	poolId uint64,
	nullifier []byte,
	senderIdentity []byte,
	recipientCommitment []byte,
	changeCommitment []byte,
	merkleRoot []byte,
	proof []byte,
	merklePathSize uint32,
) *MsgPrivateSend {
	return &MsgPrivateSend{
		Sender:             sender,
		PoolId:             poolId,
		Nullifier:          nullifier,
		SenderIdentity:    senderIdentity,
		RecipientCommitment: recipientCommitment,
		ChangeCommitment:   changeCommitment,
		MerkleRoot:         merkleRoot,
		Proof:              proof,
		MerklePathSize:     merklePathSize,
	}
}

// Route implements sdk.Msg
func (msg MsgPrivateSend) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgPrivateSend) Type() string { return TypeMsgPrivateSend }

// ValidateBasic implements sdk.HasValidateBasic
func (msg MsgPrivateSend) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrap(errortypes.ErrInvalidAddress, "invalid sender address")
	}
	if msg.PoolId == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "pool id cannot be zero")
	}
	if len(msg.Nullifier) == 0 {
		return errorsmod.Wrap(ErrInvalidNullifier, "nullifier cannot be empty")
	}
	if len(msg.SenderIdentity) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "sender identity cannot be empty")
	}
	if len(msg.RecipientCommitment) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "recipient commitment cannot be empty")
	}
	if len(msg.ChangeCommitment) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "change commitment cannot be empty")
	}
	if len(msg.MerkleRoot) == 0 {
		return errorsmod.Wrap(ErrMerkleRootMismatch, "merkle root cannot be empty")
	}
	if len(msg.Proof) == 0 {
		return errorsmod.Wrap(ErrInvalidProof, "proof cannot be empty")
	}
	if msg.MerklePathSize == 0 {
		return errorsmod.Wrap(ErrInvalidMerkleProof, "merkle path size cannot be zero")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgPrivateSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgPrivateSend) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ============================================================================
// MsgWithdrawFromCommitment - Withdrawal (amount revealed)
// ============================================================================

// NewMsgWithdrawFromCommitment creates a new MsgWithdrawFromCommitment instance
func NewMsgWithdrawFromCommitment(
	sender string,
	poolId uint64,
	commitment []byte,
	recipientAddress string,
	withdrawalAmount uint64,
	salt []byte,
	recipientIdentity []byte,
	proof []byte,
	nullifier []byte,
) *MsgWithdrawFromCommitment {
	return &MsgWithdrawFromCommitment{
		Sender:           sender,
		PoolId:           poolId,
		Commitment:       commitment,
		RecipientAddress: recipientAddress,
		WithdrawalAmount: withdrawalAmount,
		Salt:             salt,
		RecipientIdentity: recipientIdentity,
		Proof:            proof,
		Nullifier:        nullifier,
	}
}

// Route implements sdk.Msg
func (msg MsgWithdrawFromCommitment) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgWithdrawFromCommitment) Type() string { return TypeMsgWithdrawFromCommitment }

// ValidateBasic implements sdk.HasValidateBasic
func (msg MsgWithdrawFromCommitment) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrap(errortypes.ErrInvalidAddress, "invalid sender address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.RecipientAddress); err != nil {
		return errorsmod.Wrap(errortypes.ErrInvalidAddress, "invalid recipient address")
	}
	if msg.PoolId == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "pool id cannot be zero")
	}
	if len(msg.Commitment) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "commitment cannot be empty")
	}
	if msg.WithdrawalAmount == 0 {
		return errorsmod.Wrap(ErrZeroAmount, "withdrawal amount must be positive")
	}
	if len(msg.Salt) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "salt cannot be empty")
	}
	if len(msg.RecipientIdentity) == 0 {
		return errorsmod.Wrap(ErrInvalidCommitment, "recipient identity cannot be empty")
	}
	if len(msg.Proof) == 0 {
		return errorsmod.Wrap(ErrInvalidProof, "proof cannot be empty")
	}
	if len(msg.Nullifier) == 0 {
		return errorsmod.Wrap(ErrInvalidNullifier, "nullifier cannot be empty")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgWithdrawFromCommitment) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgWithdrawFromCommitment) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ============================================================================
// MsgSetVerificationKey - Set verification key (via governance)
// ============================================================================

// NewMsgSetVerificationKey creates a new MsgSetVerificationKey instance
func NewMsgSetVerificationKey(
	authority string,
	circuitType string,
	merkleDepth uint32,
	verificationKey []byte,
) *MsgSetVerificationKey {
	return &MsgSetVerificationKey{
		Authority:       authority,
		CircuitType:     circuitType,
		MerkleDepth:     merkleDepth,
		VerificationKey: verificationKey,
	}
}

// Route implements sdk.Msg
func (msg MsgSetVerificationKey) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgSetVerificationKey) Type() string { return TypeMsgSetVerificationKey }

// ValidateBasic implements sdk.HasValidateBasic
func (msg MsgSetVerificationKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(errortypes.ErrInvalidAddress, "invalid authority address")
	}
	
	// Validate circuit type
	if msg.CircuitType != CircuitTypeDeposit &&
		msg.CircuitType != CircuitTypePrivateSend &&
		msg.CircuitType != CircuitTypeWithdrawal {
		return errorsmod.Wrap(ErrInvalidProof, "invalid circuit type")
	}
	
	// For private_send circuit, merkle_depth must be > 0
	if msg.CircuitType == CircuitTypePrivateSend && msg.MerkleDepth == 0 {
		return errorsmod.Wrap(ErrInvalidMerkleProof, "merkle depth must be > 0 for private_send circuit")
	}
	
	// For deposit and withdrawal circuits, merkle_depth should be 0
	if (msg.CircuitType == CircuitTypeDeposit || msg.CircuitType == CircuitTypeWithdrawal) && msg.MerkleDepth != 0 {
		return errorsmod.Wrap(ErrInvalidMerkleProof, "merkle depth must be 0 for deposit and withdrawal circuits")
	}
	
	if len(msg.VerificationKey) == 0 {
		return errorsmod.Wrap(ErrVerificationKeyNotFound, "verification key cannot be empty")
	}
	
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgSetVerificationKey) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgSetVerificationKey) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

