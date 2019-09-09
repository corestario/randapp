package randapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/dgamingfoundation/dkglib/lib/types"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(types.MsgSendDKGData{}, "randapp/SendDKGData", nil)
}
