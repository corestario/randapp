package util

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/dgamingfoundation/randapp/x/randapp"
	"github.com/tendermint/go-amino"
)

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *amino.Codec {
	var cdc = amino.NewCodec()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	randapp.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
