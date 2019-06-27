package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSendDKGData{}, "randapp/SendDKGData", nil)
	cdc.RegisterConcrete(ed25519.PubKeyEd25519{}, "tendermint/PubKeyEd25519", nil)
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, "tendermint/PubKeySecp256k1", nil)
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
}
