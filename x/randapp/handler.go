package randapp

import (
	"fmt"

	msgs "github.com/corestario/dkglib/lib/msgs"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "randapp" type messages.
func NewHandler(keeper *Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case msgs.MsgSendDKGData:
			return handleMsgSendDKGData(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized randapp Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set name
func handleMsgSendDKGData(ctx sdk.Context, keeper *Keeper, msg msgs.MsgSendDKGData) sdk.Result {
	keeper.AddDKGData(ctx, msgs.RandDKGData{Data: msg.Data, Owner: msg.Owner})
	return sdk.Result{}
}
