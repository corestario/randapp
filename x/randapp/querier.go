package randapp

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"

	msgs "github.com/corestario/dkglib/lib/alias"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the randapp Querier
const (
	QueryDKGData = "dkgData"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper *Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryDKGData:
			return queryDKGData(ctx, path[1:], req, keeper)
		default:
			return nil, fmt.Errorf("unknown randapp query endpoint")
		}
	}
}

func queryDKGData(ctx sdk.Context, path []string, req abci.RequestQuery, keeper *Keeper) ([]byte, error) {
	if len(path) < 2 {
		return nil, fmt.Errorf("invalid query, need 2 arguments, got %d", len(path))
	}

	dataType, err := strconv.Atoi(path[0])
	if err != nil {
		return nil, fmt.Errorf("argument 1 invalid data type: %s", path[0])
	}

	roundID, err := strconv.Atoi(path[1])
	if err != nil {
		return nil, fmt.Errorf("argument 2 invalid data type: %s", path[1])
	}

	var (
		datas = keeper.GetDKGData(ctx, msgs.DKGDataType(dataType), roundID)
		buf   = bytes.NewBuffer(nil)
		enc   = gob.NewEncoder(buf)
	)
	if err := enc.Encode(datas); err != nil {
		return nil, fmt.Errorf("failed to encode response: %v", err)
	}

	return buf.Bytes(), nil
}
