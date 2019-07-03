package randapp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/randapp/x/randapp/types"
)

// NewHandler returns a handler for "randapp" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSendDKGData:
			return handleMsgSendDKGData(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized randapp Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set name
func handleMsgSendDKGData(ctx sdk.Context, keeper Keeper, msg types.MsgSendDKGData) sdk.Result {
	keeper.AddDKGData(ctx, types.DKGData{Data: msg.Data, Owner: msg.Owner}, msg.ValidatorsCount)
	return sdk.Result{}
}
