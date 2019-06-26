package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSendDKGData{}, "randapp/SendDKGData", nil)
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
}
