package randapp

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/dgamingfoundation/randapp/common"
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
	stakingKeeper staking.Keeper,
	distrKeeper distribution.Keeper,
	storeKey sdk.StoreKey,

	keyPubKeys *sdk.KVStoreKey,
	keyDeals *sdk.KVStoreKey,
	keyResponses *sdk.KVStoreKey,
	keyJustifications *sdk.KVStoreKey,
	keyCommits *sdk.KVStoreKey,
	keyComplaints *sdk.KVStoreKey,
	keyReconstructCommits *sdk.KVStoreKey,

	cdc *codec.Codec,
	cfg *config.RAServerConfig,
	msgMetr *common.MsgMetrics,
) *Keeper {
	return &Keeper{
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

func (k Keeper) AddDKGData(ctx sdk.Context, data DKGData) {
	if data.Owner.Empty() {
		return
	}

	store, err := k.getStore(ctx, data.Data.Type)
	if err != nil {
		return
	}

	var key = data.Data.Addr
	if store.Has(key) {
		return
	}

	store.Set(key, k.cdc.MustMarshalBinaryBare(data))
}

func (k Keeper) GetDKGData(ctx sdk.Context, dataType DKGDataType) []*DKGData {
	store, err := k.getStore(ctx, dataType)
	if err != nil {
		return nil
	}

	var (
		out      []*DKGData
		iterator = sdk.KVStorePrefixIterator(store, nil)
	)
	for ; iterator.Valid(); iterator.Next() {
		var data DKGData
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &data)
		out = append(out, &data)
	}

	return out
}

func (k Keeper) getStore(ctx sdk.Context, dataType DKGDataType) (sdk.KVStore, error) {
	switch dataType {
	case DKGPubKey:
		return ctx.KVStore(k.keyPubKeys), nil
	case DKGDeal:
		return ctx.KVStore(k.keyDeals), nil
	case DKGResponse:
		return ctx.KVStore(k.keyResponses), nil
	case DKGJustification:
		return ctx.KVStore(k.keyJustifications), nil
	case DKGCommits:
		return ctx.KVStore(k.keyCommits), nil
	case DKGComplaint:
		return ctx.KVStore(k.keyComplaints), nil
	case DKGReconstructCommit:
		return ctx.KVStore(k.keyReconstructCommits), nil
	default:
		return nil, fmt.Errorf("unknown message type: %d", dataType)
	}
}
