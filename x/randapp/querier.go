package randapp

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/dkglib/lib/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the randapp Querier
const (
	QueryDKGData = "dkgData"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper *Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryDKGData:
			return queryDKGData(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown randapp query endpoint")
		}
	}
}

func queryDKGData(ctx sdk.Context, path []string, req abci.RequestQuery, keeper *Keeper) (res []byte, err sdk.Error) {
	fmt.Println("QUERYDKG:", path[0])
	dataType, err1 := strconv.Atoi(path[0])
	if err1 != nil {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("invalid data type: %s", path[0]))
	}

	var (
		datas = keeper.GetDKGData(ctx, types.DKGDataType(dataType))
		buf   = bytes.NewBuffer(nil)
		enc   = gob.NewEncoder(buf)
	)
	fmt.Println("QUERYDKGDATA:", len(datas))
	if err := enc.Encode(datas); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to encode response: %v", err))
	}

	return buf.Bytes(), nil
}
