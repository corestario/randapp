package randapp

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	appTypes "github.com/dgamingfoundation/randapp/x/randapp/types"
	"github.com/tendermint/tendermint/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
// for the various parts of the state machine.
type Keeper struct {
	coinKeeper            bank.Keeper
	keyPubKeys            *sdk.KVStoreKey
	keyDeals              *sdk.KVStoreKey
	keyResponses          *sdk.KVStoreKey
	keyJustifications     *sdk.KVStoreKey
	keyCommits            *sdk.KVStoreKey
	keyComplaints         *sdk.KVStoreKey
	keyReconstructCommits *sdk.KVStoreKey
	cdc                   *codec.Codec // The wire codec for binary encoding/decoding.
}

func NewKeeper(
	coinKeeper bank.Keeper,
	keyPubKeys *sdk.KVStoreKey,
	keyDeals *sdk.KVStoreKey,
	keyResponses *sdk.KVStoreKey,
	keyJustifications *sdk.KVStoreKey,
	keyCommits *sdk.KVStoreKey,
	keyComplaints *sdk.KVStoreKey,
	keyReconstructCommits *sdk.KVStoreKey,
	cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper:            coinKeeper,
		keyPubKeys:            keyPubKeys,
		keyDeals:              keyDeals,
		keyResponses:          keyResponses,
		keyJustifications:     keyJustifications,
		keyCommits:            keyCommits,
		keyComplaints:         keyComplaints,
		keyReconstructCommits: keyReconstructCommits,
		cdc:                   cdc,
	}
}

func (k Keeper) AddDKGData(ctx sdk.Context, data appTypes.DKGData) {
	fmt.Println("ADD DKG DATA!")
	if data.Owner.Empty() {
		return
	}

	store, err := k.getStore(ctx, data.Data.Type)
	if err != nil {
		return
	}

	var bas = data.Data.Addr
	for i := 0; i < 4; i++ {
		key := append(bas, byte(i))
		if !store.Has(key) {
			store.Set(key, k.cdc.MustMarshalBinaryBare(data))
			return
		}
	}

}

func (k Keeper) GetDKGData(ctx sdk.Context, dataType types.DKGDataType) []*types.DKGData {
	store, err := k.getStore(ctx, dataType)
	if err != nil {
		return nil
	}

	var (
		out      []*types.DKGData
		iterator = sdk.KVStorePrefixIterator(store, nil)
	)
	for ; iterator.Valid(); iterator.Next() {
		var data appTypes.DKGData
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &data)
		out = append(out, data.Data)
	}
	return out
}

func (k Keeper) getStore(ctx sdk.Context, dataType types.DKGDataType) (sdk.KVStore, error) {
	switch dataType {
	case types.DKGPubKey:
		return ctx.KVStore(k.keyPubKeys), nil
	case types.DKGDeal:
		return ctx.KVStore(k.keyDeals), nil
	case types.DKGResponse:
		return ctx.KVStore(k.keyResponses), nil
	case types.DKGJustification:
		return ctx.KVStore(k.keyJustifications), nil
	case types.DKGCommits:
		return ctx.KVStore(k.keyCommits), nil
	case types.DKGComplaint:
		return ctx.KVStore(k.keyComplaints), nil
	case types.DKGReconstructCommit:
		return ctx.KVStore(k.keyReconstructCommits), nil
	default:
		return nil, fmt.Errorf("unknown message type: %d", dataType)
	}
}
