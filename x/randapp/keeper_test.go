package randapp_test

import (
	"testing"

	"github.com/corestario/dkglib/lib/alias"
	"github.com/corestario/dkglib/lib/msgs"
	"github.com/corestario/randapp/x/randapp/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func createTestKeeper() (*app.RandApp, sdk.Context) {
	testApp := app.Setup(false)
	ctx := testApp.NewContext(false, abci.Header{})

	return testApp, ctx
}

func TestKeeper_AddDKGData(t *testing.T) {
	testApp, ctx := createTestKeeper()

	req := require.New(t)
	t.Run("test_empty_owner", func(t *testing.T) {
		err := testApp.GetRandappKeeper().AddDKGData(ctx, msgs.MsgSendDKGData{Data: &alias.DKGData{
			Type:    alias.DKGDeal,
			RoundID: 1,
		}})
		req.Error(err)
	})

	t.Run("test_invalid_type", func(t *testing.T) {
		err := testApp.GetRandappKeeper().AddDKGData(ctx, msgs.MsgSendDKGData{Data: &alias.DKGData{
			Type:    666,
			RoundID: 1,
		},
			Owner: sdk.AccAddress("test_address"),
		})
		req.Error(err)
	})

	t.Run("test_ok", func(t *testing.T) {
		err := testApp.GetRandappKeeper().AddDKGData(ctx, msgs.MsgSendDKGData{Data: &alias.DKGData{
			Type:    alias.DKGDeal,
			RoundID: 1,
		},
			Owner: sdk.AccAddress("test_address"),
		})
		req.NoError(err)
	})
}
