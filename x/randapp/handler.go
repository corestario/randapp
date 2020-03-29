package randapp

import (
	"fmt"

	msgs "github.com/corestario/dkglib/lib/msgs"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "randapp" type messages.
func NewHandler(keeper *Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case msgs.MsgSendDKGData:
			return handleMsgSendDKGData(ctx, keeper, msg)
		default:
			return nil, fmt.Errorf("unrecognized randapp Msg type: %v", msg.Type())
		}
	}
}

// Handle a message to set name
func handleMsgSendDKGData(ctx sdk.Context, keeper *Keeper, msg msgs.MsgSendDKGData) (*sdk.Result, error) {
	if err := keeper.AddDKGData(ctx, msgs.MsgSendDKGData{Data: msg.Data, Owner: msg.Owner}); err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent("EventAdDKGDataFailed",
			sdk.NewAttribute("error", err.Error())))
		return &sdk.Result{Log: err.Error()}, err
	}
	return &sdk.Result{}, nil
}
