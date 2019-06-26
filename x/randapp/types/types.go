package types

import (
	"fmt"

	"github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DKGData struct {
	Data  *types.DKGData `json:"data"`
	Owner sdk.AccAddress `json:"owner"`
}

func (m DKGData) String() string {
	return fmt.Sprintf("Data: %+v, Owner: %s", m.Data, m.Owner.String())
}
