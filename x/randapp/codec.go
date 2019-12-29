package randapp

import (
	types "github.com/corestario/dkglib/lib/msgs"
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(types.MsgSendDKGData{}, "randapp/SendDKGData", nil)
}
