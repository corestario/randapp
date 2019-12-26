package randapp

import (
	"fmt"

	types "github.com/corestario/dkglib/lib/alias"
	msgs "github.com/corestario/dkglib/lib/msgs"
	"github.com/corestario/randapp/common"
	"github.com/corestario/randapp/x/randapp/config"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	pl "github.com/prometheus/common/log"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
// for the various parts of the state machine.
type Keeper struct {
	coinKeeper    bank.Keeper
	stakingKeeper staking.Keeper
	distrKeeper   distribution.Keeper
	storeKey      sdk.StoreKey // Unexposed key to access store from sdk.Context

	keyPubKeys            *sdk.KVStoreKey
	keyDeals              *sdk.KVStoreKey
	keyResponses          *sdk.KVStoreKey
	keyJustifications     *sdk.KVStoreKey
	keyCommits            *sdk.KVStoreKey
	keyComplaints         *sdk.KVStoreKey
	keyReconstructCommits *sdk.KVStoreKey

	cdc     *codec.Codec // The wire codec for binary encoding/decoding.
	config  *config.RAServerConfig
	msgMetr *common.MsgMetrics
}

func NewKeeper(
	coinKeeper bank.Keeper,
	stakingKeeper staking.Keeper,
	distrKeeper distribution.Keeper,

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
		stakingKeeper:         stakingKeeper,
		distrKeeper:           distrKeeper,
		keyPubKeys:            keyPubKeys,
		keyDeals:              keyDeals,
		keyResponses:          keyResponses,
		keyJustifications:     keyJustifications,
		keyCommits:            keyCommits,
		keyComplaints:         keyComplaints,
		keyReconstructCommits: keyReconstructCommits,
		cdc:                   cdc,
		config:                cfg,
		msgMetr:               msgMetr,
	}
}

func (k *Keeper) increaseCounter(labels ...string) {
	counter, err := k.msgMetr.NumMsgs.GetMetricWithLabelValues(labels...)
	if err != nil {
		pl.Errorf("get metrics with label values error: %v", err)
		return
	}
	counter.Inc()
}

func makeKey(roundID int, count int) []byte {
	return []byte(makePrefix(roundID) + fmt.Sprintf("_%d", count))
}

func makePrefix(roundID int) string {
	return fmt.Sprintf("round_%d", roundID)
}

func getMax(validatorCount int, dataType types.DKGDataType) int {
	res := 1
	vc := validatorCount - 1
	switch dataType {
	case types.DKGPubKey:
	case types.DKGDeal:
		res = vc
	case types.DKGResponse:
		res = vc
	case types.DKGJustification:
		res = vc * vc
	case types.DKGCommits:
	case types.DKGComplaint:
	case types.DKGReconstructCommit:
	default:
		res = 0
	}
	return res * validatorCount
}

func (k Keeper) AddDKGData(ctx sdk.Context, data msgs.RandDKGData) {
	if data.Owner.Empty() {
		return
	}

	store, err := k.getStore(ctx, data.Data.Type)
	if err != nil {
		return
	}

	var bas = data.Data.Addr
	for i := 0; i < getMax(4, data.Data.Type); i++ {
		key := append(makeKey(data.Data.RoundID, i), bas...)
		if !store.Has(key) {
			store.Set(key, k.cdc.MustMarshalBinaryBare(data))
			return
		}
	}
}

func (k Keeper) GetDKGData(ctx sdk.Context, dataType DKGDataType) []*msgs.RandDKGData {
	fmt.Printf("GETDKGDATA: %+v \t", dataType)

	store, err := k.getStore(ctx, dataType)
	if err != nil {
		return nil
	}

	var (
		out      []*msgs.RandDKGData
		iterator = sdk.KVStorePrefixIterator(store, nil)
	)
	for ; iterator.Valid(); iterator.Next() {
		var data msgs.RandDKGData
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &data)
		out = append(out, &data)
	}

	fmt.Println("DKGLEN = ", len(out))

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
