package randapp

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	appTypes "github.com/dgamingfoundation/randapp/x/randapp/types"
	"github.com/tendermint/tendermint/types"
	"log"
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

func makeKey(roundID int) []byte {
	return []byte(makePrefix(roundID))
}

func makePrefix(roundID int) string {
	return fmt.Sprintf("round_%d", roundID)
}

func getMax(validatorCount int, dataType types.DKGDataType) int {
	res := 1
	switch dataType {
	case types.DKGPubKey:
	case types.DKGDeal:
		res = validatorCount
	case types.DKGResponse:
		res = validatorCount - 1
	case types.DKGJustification:
		p := validatorCount - 1
		res = p * p
	case types.DKGCommits:
	case types.DKGComplaint:
	case types.DKGReconstructCommit:
	default:
		res = 0
	}
	return res * validatorCount
}

func createStore(validatorCount int, dataType types.DKGDataType) appTypes.MessageStore {
	var mStore appTypes.MessageStore
	switch dataType {
	case types.DKGPubKey, types.DKGReconstructCommit, types.DKGComplaint, types.DKGCommits, types.DKGDeal:
		mStore = appTypes.NewMessageStore(1)
	case types.DKGResponse:
		mStore = appTypes.NewMessageStore(validatorCount - 1)
	case types.DKGJustification:
		p := validatorCount - 1
		mStore = appTypes.NewMessageStore(p * p)
	}
	log.Println("STORE CREATED")
	return mStore
}

func (k Keeper) AddDKGData(ctx sdk.Context, data appTypes.DKGData, validatorCount int) {
	var (
		mStore appTypes.MessageStore
	)
	if data.Owner.Empty() {
		return
	}

	store, err := k.getStore(ctx, data.Data.Type)
	if err != nil {
		return
	}

	key := makeKey(data.Data.RoundID)

	mStoreBytes := store.Get(key)
	if mStoreBytes == nil {
		mStore = createStore(validatorCount, data.Data.Type)
	} else {
		k.cdc.MustUnmarshalBinaryBare(mStoreBytes, &mStore)
	}

	mStore.Add(data.Owner.String(), k.cdc.MustMarshalBinaryBare(data))

	store.Set(key, k.cdc.MustMarshalBinaryBare(mStore))

	log.Println("ADD DATA:", data.Data.Type)
}

func (k Keeper) GetDKGData(ctx sdk.Context, dataType types.DKGDataType, roundID int) []*types.DKGData {
	store, err := k.getStore(ctx, dataType)
	if err != nil {
		return nil
	}

	var (
		out    []*types.DKGData
		mStore appTypes.MessageStore
		data   types.DKGData
	)

	mStoreBytes := store.Get(makeKey(roundID))
	if mStoreBytes == nil {
		log.Println("MessageStore doesn't exist!!!!!!!!!")
		return out
	}
	k.cdc.MustUnmarshalBinaryBare(mStoreBytes, &mStore)

	log.Println("GET DATA:", dataType, mStore.GetMessagesCount())

	if mStore.GetMessagesCount() < getMax(4, dataType) {
		return nil
	}

	for _, peerMsgs := range mStore.GetAll() {
		for _, msg := range peerMsgs {
			k.cdc.MustUnmarshalBinaryBare(msg, &data)
			out = append(out, &data)
		}
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
