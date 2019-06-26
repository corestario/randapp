package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSendDKGData{}, "randapp/SendDKGData", nil)
	cdc.RegisterConcrete(ed25519.PubKeyEd25519{}, "tendermint/PubKeyEd25519", nil)
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
}
