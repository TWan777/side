package types

import (
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgWithdraw = "withdraw"

var _ sdk.Msg = &MsgWithdraw{}

func NewMsgWithdraw(
	sender, poolID, receiver string,
	share sdk.Coin,
) *MsgWithdraw {
	return &MsgWithdraw{
		Sender:   sender,
		Receiver: receiver,
		PoolId:   poolID,
		Share:    share,
	}
}

func (msg *MsgWithdraw) Route() string {
	return RouterKey
}

func (msg *MsgWithdraw) Type() string {
	return TypeMsgWithdraw
}

func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdraw) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid Sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	if strings.TrimSpace(msg.PoolId) == "" {
		return sdkerrors.Wrap(ErrInvalidPoolID, "pool id cannot be empty")
	}

	if msg.Share.Amount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidTokenAmount, "share amount cannot be zero")
	}
	return nil
}
